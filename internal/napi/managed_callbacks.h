#ifndef NAPI_GO_MANAGED_CALLBACKS_H
#define NAPI_GO_MANAGED_CALLBACKS_H

#include <stdint.h>
#include <stdlib.h>
#include <node/node_api.h>

typedef struct { uintptr_t handle; } napi_go_token;

napi_go_token *napi_go_token_new(uintptr_t handle);
void napi_go_token_delete(napi_go_token *token);
napi_value napi_go_method(napi_env env, napi_callback_info info);
napi_value napi_go_getter(napi_env env, napi_callback_info info);
napi_value napi_go_setter(napi_env env, napi_callback_info info);
void napi_go_callback_cleanup(void *data);
void napi_go_callback_finalize(node_api_basic_env env, void *data, void *hint);
void napi_go_tsfn_call(napi_env env, napi_value fn, void *context, void *data);
void napi_go_tsfn_finalize(napi_env env, void *data, void *hint);
void napi_go_wrap_finalize(node_api_basic_env env, void *data, void *hint);
napi_status napi_go_wrap(napi_env env, napi_value object, napi_go_token *token);
napi_status napi_go_unwrap(napi_env env, napi_value object, napi_go_token **token, int remove);

#endif
