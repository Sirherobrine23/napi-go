#ifndef __ENTRY_ENTRY_H__
#define __ENTRY_ENTRY_H__

#include <node/node_api.h>

#ifdef __cplusplus
extern "C" {
#endif /* __cplusplus */

// initializeModule is a N-API module initialization function.
// initializeModule is suitable for use as a napi_addon_register_func.
extern napi_value initializeModule(
  napi_env    env,
  napi_value  exports
);

#ifdef __cplusplus
}
#endif /* __cplusplus */

#endif /* __ENTRY_ENTRY_H__ */
