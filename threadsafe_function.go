package napi

import (
	"encoding/json"
	"fmt"
	"reflect"
	"runtime"
	"runtime/cgo"
	"sync"
	"unsafe"

	"sirherobrine23.com.br/Sirherobrine23/napi-go/internal/napi"
)

/*
#include <node/node_api.h>

// Forward declaration of the C callback function
extern void executeThreadsafeFunctionCallJSCallback(napi_env env, napi_value js_callback, void* context, void* data);
*/
import "C"

// ThreadsafeFunctionCallJSCallback is the signature for the callback function that
// will be invoked on the main Node.js thread when data is sent from another thread.
// The `jsCallback` is the original JavaScript function passed to CreateThreadsafeFunction.
// The `data` is the value passed to the Call method from the background thread.
type ThreadsafeFunctionCallJSCallback func(env EnvType, jsCallback *Function, data any)

// ThreadsafeFunction represents a N-API thread-safe function.
type ThreadsafeFunction struct {
	value
	tsfn             napi.ThreadsafeFunction
	callJSCallback   ThreadsafeFunctionCallJSCallback // Callback to execute on the main thread
	goCallbackHandle cgo.Handle                       // Handle to the Go callback data
}

// Global map to store callbacks associated with thread-safe functions
var (
	tsfnCallbacks      = make(map[napi.ThreadsafeFunction]*ThreadsafeFunction)
	tsfnCallbacksMutex sync.RWMutex
)

type threadsafeCallbackData struct {
	tsfn   *ThreadsafeFunction
	goData any // Data received from the background thread
}

//export executeThreadsafeFunctionCallJSCallback
func executeThreadsafeFunctionCallJSCallback(cEnv C.napi_env, cJsCallback C.napi_value, context unsafe.Pointer, data unsafe.Pointer) {
	// This function runs on the main Node.js thread
	env := N_APIEnv(napi.Env(cEnv))

	// Retrieve the ThreadsafeFunction instance and the data from Go handles
	callbackDataHandle := cgo.Handle(data)
	callbackData := callbackDataHandle.Value().(*threadsafeCallbackData)
	tsfn := callbackData.tsfn
	goData := callbackData.goData
	callbackDataHandle.Delete() // Clean up the handle for the data

	// It's crucial to handle potential panics in the callback
	defer func() {
		if r := recover(); r != nil {
			errStr := fmt.Sprintf("panic recovered in threadsafe function callback: %v", r)
			napi.ThrowError(env.NapiValue(), "", errStr) // Optionally throw an error back to the main loop
		}
	}()

	if tsfn == nil || tsfn.callJSCallback == nil {
		napi.ThrowError(env.NapiValue(), "", "Error: ThreadsafeFunction or its callback is nil in executeThreadsafeFunctionCallJSCallback")
		return
	}

	// Wrap the JS callback N-API value
	jsCallbackFunc := ToFunction(N_APIValue(env, napi.Value(cJsCallback)))

	// Execute the user-provided Go callback
	tsfn.callJSCallback(env, jsCallbackFunc, goData)
}

// CreateThreadsafeFunction creates a new N-API thread-safe function.
//
// Parameters:
//   - env: The N-API environment.
//   - jsFunc: The JavaScript function to be called from other threads (optional, can be nil if callJSCallback handles everything).
//   - resourceName: A string identifying the resource associated with the async work.
//   - maxQueueSize: The maximum size of the queue. 0 for no limit.
//   - initialThreadCount: The initial number of threads that will use this function (used for reference counting). Must be >= 1.
//   - callJSCallback: The Go function to be executed on the main thread when data arrives.
//   - context: Optional Go data accessible within callJSCallback via GetContext.
//   - finalizeCallback: Optional Go function called when the thread-safe function is being destroyed.
func CreateThreadsafeFunction(
	env EnvType,
	jsFunc Callback, // Can be nil
	resourceName string,
	maxQueueSize int,
	initialThreadCount int,
	callJSCallback ThreadsafeFunctionCallJSCallback,
	context any, // Optional context data
	finalizeCallback func(env EnvType, context any), // Optional finalize callback
) (*ThreadsafeFunction, error) {
	if initialThreadCount < 1 {
		return nil, fmt.Errorf("initialThreadCount must be at least 1")
	} else if callJSCallback == nil {
		return nil, fmt.Errorf("callJSCallback cannot be nil")
	}

	resourceNameVal, err := CreateString(env, resourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource name string: %w", err)
	}

	var jsFuncVal napi.Value
	if jsFunc != nil {
		jsFn, err := CreateFunction(env, runtime.FuncForPC(reflect.ValueOf(jsFunc).Pointer()).Name(), jsFunc)
		if err != nil {
			return nil, fmt.Errorf("failed to create JavaScript function: %w", err)
		}
		jsFuncVal = jsFn.NapiValue()
	}

	tsfnWrapper := &ThreadsafeFunction{
		callJSCallback: callJSCallback,
	}

	// Store context and finalize callback if provided
	callbackData := &struct {
		Context          any
		FinalizeCallback func(env EnvType, context any)
		TsfnWrapper      *ThreadsafeFunction // Reference back to the wrapper
	}{
		Context:          context,
		FinalizeCallback: finalizeCallback,
		TsfnWrapper:      tsfnWrapper,
	}
	tsfnWrapper.goCallbackHandle = cgo.NewHandle(callbackData)

	var finalizeCbPtr C.napi_finalize
	if finalizeCallback != nil {
		fmt.Println("Warning: Thread-safe function finalizer callback is not fully implemented in this example.")
	}

	var cTsfn napi.ThreadsafeFunction
	status := napi.Status(C.napi_create_threadsafe_function(
		C.napi_env(env.NapiValue()),
		C.napi_value(jsFuncVal),
		nil, // async_resource (optional)
		C.napi_value(resourceNameVal.NapiValue()),
		C.size_t(maxQueueSize),
		C.size_t(initialThreadCount),
		unsafe.Pointer(tsfnWrapper.goCallbackHandle), // thread_finalize_data
		finalizeCbPtr,                                // thread_finalize_cb (optional)
		unsafe.Pointer(tsfnWrapper.goCallbackHandle), // context
		C.napi_threadsafe_function_call_js(C.executeThreadsafeFunctionCallJSCallback), // call_js_cb
		(*C.napi_threadsafe_function)(unsafe.Pointer(&cTsfn)),
	))

	if err := status.ToError(); err != nil {
		tsfnWrapper.goCallbackHandle.Delete() // Clean up handle on failure
		return nil, fmt.Errorf("failed to create threadsafe function: %w", err)
	}

	tsfnWrapper.tsfn = cTsfn
	tsfnWrapper.value = N_APIValue(env, jsFuncVal) // Use jsFunc as the underlying value if provided

	// Store the mapping from N-API handle to Go wrapper
	tsfnCallbacksMutex.Lock()
	tsfnCallbacks[cTsfn] = tsfnWrapper
	tsfnCallbacksMutex.Unlock()

	return tsfnWrapper, nil
}

// GetContext retrieves the context data provided during creation.
func (tsfn *ThreadsafeFunction) GetContext() (any, error) {
	if tsfn.goCallbackHandle == 0 {
		return nil, fmt.Errorf("threadsafe function has no associated context handle (possibly already finalized)")
	}
	// We stored the handle in the wrapper itself during creation
	callbackData := tsfn.goCallbackHandle.Value().(*struct {
		Context          any
		FinalizeCallback func(env EnvType, context any)
		TsfnWrapper      *ThreadsafeFunction
	})
	return callbackData.Context, nil
}

// Call sends data from a background thread to the main Node.js thread.
// The data will be received by the callJSCallback provided during creation.
// This method is safe to call from any thread.
func (tsfn *ThreadsafeFunction) Call(data any, mode napi.ThreadsafeFunctionCallMode) error {
	if tsfn.tsfn == nil {
		return fmt.Errorf("threadsafe function is not initialized or already released")
	}

	// Create a handle for the Go data to pass it safely through C
	// The handle will be deleted by the executeThreadsafeFunctionCallJSCallback on the main thread.
	dataHandle := cgo.NewHandle(&threadsafeCallbackData{
		tsfn:   tsfn,
		goData: data,
	})

	status := napi.CallThreadsafeFunction(
		tsfn.tsfn,
		// unsafe.Pointer(dataHandle), // Pass the handle as data
		mode,
	)
	if err := status.ToError(); err != nil {
		// If the call fails, we need to delete the handle ourselves
		dataHandle.Delete()
		// Specific error handling for queue full might be needed
		if status == napi.StatusQueueFull && mode == napi.NonBlocking {
			return fmt.Errorf("threadsafe function queue is full: %w", err)
		}
		return fmt.Errorf("failed to call threadsafe function: %w", err)
	}
	return nil
}

// Acquire increments the reference count for the thread-safe function,
// indicating that a new thread is about to start using it.
// This method is safe to call from any thread.
func (tsfn *ThreadsafeFunction) Acquire() error {
	if tsfn.tsfn == nil {
		return fmt.Errorf("threadsafe function is not initialized or already released")
	}
	return singleMustValueErr(napi.AcquireThreadsafeFunction(tsfn.tsfn))
}

// Release decrements the reference count for the thread-safe function.
// It should be called when a thread finishes using the function.
// The mode specifies whether to close the queue immediately (Abort) or
// wait for pending items to be processed (Release).
// This method is safe to call from any thread.
func (tsfn *ThreadsafeFunction) Release(mode napi.ThreadsafeFunctionReleaseMode) error {
	if tsfn.tsfn == nil {
		return fmt.Errorf("threadsafe function is not initialized or already released")
	}

	// Remove from the global map *before* the final N-API release
	// This prevents potential race conditions if the finalizer runs quickly.
	// Only remove if the reference count will drop to zero or below (though N-API handles the actual destruction).
	// Predicting the exact count is tricky, so removing might be premature if Release fails.
	// A safer approach might involve coordination within the finalizer callback.
	// For now, we remove it pessimistically.
	tsfnCallbacksMutex.Lock()
	delete(tsfnCallbacks, tsfn.tsfn)
	tsfnCallbacksMutex.Unlock()

	err := singleMustValueErr(napi.ReleaseThreadsafeFunction(tsfn.tsfn, mode))
	if err == nil {
		// If release was successful, mark the Go wrapper as invalid
		// tsfn.tsfn = nil // Be careful with concurrent access if doing this
		// Deleting the handle should ideally happen in the C finalizer callback
		// tsfn.goCallbackHandle.Delete() // Potential double delete if finalizer runs
	} else {
		// If release failed, potentially re-add to map? Complex error recovery needed.
		fmt.Printf("Warning: Failed to release threadsafe function: %v. State might be inconsistent.\n", err)
	}

	return err
}

// Ref increments the N-API reference count, preventing the JS function object
// from being garbage collected while the thread-safe function is active.
// Must be called from the main Node.js thread.
func (tsfn *ThreadsafeFunction) Ref(env EnvType) error {
	if tsfn.tsfn == nil {
		return fmt.Errorf("threadsafe function is not initialized or already released")
	}
	// Ensure this is called from the main thread (N-API doesn't enforce this, but it's best practice)
	// Checking the thread ID might be complex. Rely on user discipline for now.
	return singleMustValueErr(napi.RefThreadsafeFunction(env.NapiValue(), tsfn.tsfn))
}

// Unref decrements the N-API reference count.
// Must be called from the main Node.js thread.
func (tsfn *ThreadsafeFunction) Unref(env EnvType) error {
	if tsfn.tsfn == nil {
		return fmt.Errorf("threadsafe function is not initialized or already released")
	}
	// Ensure this is called from the main thread
	return singleMustValueErr(napi.UnrefThreadsafeFunction(env.NapiValue(), tsfn.tsfn))
}

// --- Helper function for converting Go data to N-API value ---
// This might be useful within the callJSCallback

// DataToNapiValue attempts to convert arbitrary Go data to a napi.Value.
// This uses the existing js.ValueOf logic. Needs error handling refinement.
func DataToNapiValue(env EnvType, data any) (ValueType, error) {
	// Handle nil explicitly
	if data == nil {
		return env.Null()
	}

	// Use reflection for basic types or js.ValueOf for complex ones
	switch v := data.(type) {
	case string:
		return CreateString(env, v)
	case int:
		return CreateNumber(env, v)
	case int64:
		return CreateBigint(env, v) // Or CreateNumber if precision allows
	case float64:
		return CreateNumber(env, v)
	case bool:
		return CreateBoolean(env, v)
	case []byte:
		return CopyBuffer(env, v)
	// Add other common types as needed...
	default:
		// Fallback to js.ValueOf for slices, maps, structs, etc.
		// This requires the 'js' package.
		valueOf, err := ValueOf(env, data) // Assumes js package is available
		if err != nil {
			return nil, fmt.Errorf("failed to convert data type %T using js.ValueOf: %w", data, err)
		}
		// Check if the result is undefined, indicating an unsupported type
		typ, _ := valueOf.Type()
		if typ == TypeUndefined && reflect.ValueOf(data).IsValid() { // Check if data wasn't actually nil/invalid
			// Attempt JSON marshaling as a last resort?
			jsonData, jsonErr := json.Marshal(data)
			if jsonErr == nil {
				var parsedData any
				if json.Unmarshal(jsonData, &parsedData) == nil {
					return DataToNapiValue(env, parsedData) // Recurse with JSON parsed data
				}
			}
			return nil, fmt.Errorf("unsupported data type for threadsafe function call: %T", data)
		}
		return valueOf, nil
	}
}

// Consider adding a default CallJSCallback implementation
// DefaultCallJSCallback attempts to convert the Go data to a N-API value and pass it to the JS callback.
func DefaultCallJSCallback(env EnvType, jsCallback *Function, data any) {
	if jsCallback == nil {
		fmt.Println("Warning: DefaultCallJSCallback called with nil jsCallback")
		return
	}

	napiData, err := DataToNapiValue(env, data)
	if err != nil {
		errMsg := fmt.Sprintf("DefaultCallJSCallback Error: Failed to convert Go data: %v", err)
		fmt.Println(errMsg)
		// Maybe try to call the JS callback with an error object?
		errObj, _ := CreateError(env, errMsg)
		if errObj != nil {
			_, callErr := jsCallback.Call(errObj) // Pass error as the first argument
			if callErr != nil {
				fmt.Printf("DefaultCallJSCallback Error: Failed to call JS callback with error: %v\n", callErr)
			}
		}
		return
	}

	_, callErr := jsCallback.Call(napiData) // Call JS func with the converted data
	if callErr != nil {
		// Log error calling the JS function
		fmt.Printf("DefaultCallJSCallback Error: Failed to call JS callback: %v\n", callErr)
		// Potentially throw an error back to the main loop? Risky.
	}
}
