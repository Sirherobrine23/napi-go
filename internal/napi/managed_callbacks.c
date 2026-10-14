#include "managed_callbacks.h"

extern napi_value executeManagedCallback(napi_env, napi_callback_info, uintptr_t, int);
extern void deleteManagedCallback(uintptr_t);
extern void executeManagedThreadsafeCall(napi_env, napi_value, uintptr_t, uintptr_t);
extern void finalizeManagedThreadsafeFunction(napi_env, uintptr_t);
extern void finalizeManagedWrap(uintptr_t);

napi_go_token *napi_go_token_new(uintptr_t handle) {
  napi_go_token *token = malloc(sizeof(*token));
  if (token) token->handle = handle;
  return token;
}

void napi_go_token_delete(napi_go_token *token) { free(token); }

static napi_value dispatch(napi_env env, napi_callback_info info, int kind) {
  void *data = NULL;
  if (napi_get_cb_info(env, info, NULL, NULL, NULL, &data) != napi_ok || !data)
    return NULL;
  return executeManagedCallback(env, info, ((napi_go_token *)data)->handle, kind);
}

napi_value napi_go_method(napi_env env, napi_callback_info info) { return dispatch(env, info, 0); }
napi_value napi_go_getter(napi_env env, napi_callback_info info) { return dispatch(env, info, 1); }
napi_value napi_go_setter(napi_env env, napi_callback_info info) { return dispatch(env, info, 2); }

void napi_go_callback_cleanup(void *data) {
  napi_go_token *token = data;
  deleteManagedCallback(token->handle);
  free(token);
}

void napi_go_callback_finalize(node_api_basic_env env, void *data, void *hint) {
  napi_go_callback_cleanup(data);
}

void napi_go_tsfn_call(napi_env env, napi_value fn, void *context, void *data) {
  napi_go_token *token = data;
  uintptr_t handle = token ? token->handle : 0;
  free(token);
  // Node may finalize the function before disposing the abandoned queue.
  // During that disposal context is no longer live and must not be read.
  executeManagedThreadsafeCall(env, fn, env ? ((napi_go_token *)context)->handle : 0, handle);
}

void napi_go_tsfn_finalize(napi_env env, void *data, void *hint) {
  napi_go_token *token = data;
  uintptr_t handle = token->handle;
  free(token);
  finalizeManagedThreadsafeFunction(env, handle);
}

// Distinguish napi-go wraps from arbitrary native pointers owned by other addons.
static const napi_type_tag wrap_tag = { UINT64_C(0x65b43f2c8da09e17), UINT64_C(0xa948d13b2076fc85) };

void napi_go_wrap_finalize(node_api_basic_env env, void *data, void *hint) {
  napi_go_token *token = data;
  uintptr_t handle = token->handle;
  free(token);
  finalizeManagedWrap(handle);
}

napi_status napi_go_wrap(napi_env env, napi_value object, napi_go_token *token) {
  napi_status status = napi_wrap(env, object, token, napi_go_wrap_finalize, NULL, NULL);
  if (status != napi_ok) return status;
  bool tagged = false;
  status = napi_check_object_type_tag(env, object, &wrap_tag, &tagged);
  if (status == napi_ok && !tagged) status = napi_type_tag_object(env, object, &wrap_tag);
  if (status != napi_ok) {
    void *removed = NULL;
    napi_remove_wrap(env, object, &removed);
  }
  return status;
}

napi_status napi_go_unwrap(napi_env env, napi_value object, napi_go_token **token, int remove) {
  bool tagged = false;
  napi_status status = napi_check_object_type_tag(env, object, &wrap_tag, &tagged);
  if (status != napi_ok) return status;
  if (!tagged) return napi_invalid_arg;
  return remove ? napi_remove_wrap(env, object, (void **)token)
                : napi_unwrap(env, object, (void **)token);
}
