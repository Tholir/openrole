// Package plugin provides the WASM plugin runtime for openrole using GopherJS.
// Plugins are compiled Go modules that run in the browser via WebAssembly.
package plugin

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"syscall/js"
)

// WASM represents a loaded WebAssembly plugin instance.
type WASM struct {
	instance js.Value
	module   js.Value
	exports  js.Value

	// plugin instance metadata
	id      string
	loadedAt int64

	mu sync.RWMutex
}

// Global plugin registry for managing multiple WASM instances.
var (
	registry     = make(map[string]*WASM)
	registryMu   sync.RWMutex
	moduleCache  = make(map[string]js.Value)
	moduleCacheMu sync.RWMutex
)

// NewWASM loads a WASM module from the given byte slice and wraps it.
// The jsCtx is the JavaScript global context used for WASM execution.
func NewWASM(ctx context.Context, id string, wasmBytes []byte, jsCtx js.Value) (*WASM, error) {
	registryMu.RLock()
	if existing, ok := registry[id]; ok {
		registryMu.RUnlock()
		return existing, nil
	}
	registryMu.RUnlock()

	// Compile the WASM module using GopherJS runtime.
	module, err := compileWASM(ctx, wasmBytes, jsCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to compile WASM module %q: %w", id, err)
	}

	// Instantiate the module.
	instance, err := instantiateModule(ctx, module, jsCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate WASM module %q: %w", id, err)
	}

	wasm := &WASM{
		instance:  instance,
		module:    module,
		exports:   instance.Get("exports"),
		id:        id,
		loadedAt:  now(),
	}

	registryMu.Lock()
	registry[id] = wasm
	registryMu.Unlock()

	slog.Info("WASM plugin loaded", "id", id)
	return wasm, nil
}

// compileWASM compiles Go-compiled WASM bytes using the GopherJS runtime.
// The jsCtx provides the JavaScript environment for compilation.
func compileWASM(ctx context.Context, wasmBytes []byte, jsCtx js.Value) (js.Value, error) {
	// GopherJS provides a global __go傑 compiler that handles WASM compilation.
	// The runtime expects WASM as an ArrayBuffer.
	compiler := jsCtx.Get("__go杰")
	if !compiler.Truthy() {
		return js.Value{}, fmt.Errorf("GopherJS runtime not found; ensure gopherjs.js is loaded")
	}

	// Convert byte slice to JavaScript Uint8Array.
	jsBytes := jsarrayFromBytes(wasmBytes, jsCtx)

	// Call the GopherJS compiler with the WASM bytes.
	result := compiler.Invoke(jsBytes)
	if isPromise(result) {
		result = awaitPromise(ctx, result)
	}

	if result.IsNull() || result.IsUndefined() {
		return js.Value{}, fmt.Errorf("GopherJS compiler returned null")
	}

	return result, nil
}

// instantiateModule creates a new WASM instance from a compiled module.
// GopherJS uses a specific importObject pattern for Go runtime.
func instantiateModule(ctx context.Context, module js.Value, jsCtx js.Value) (js.Value, error) {
	// Build the import object for Go WASM runtime.
	importObj := buildImportObject(jsCtx)

	// WebAssembly.instantiate expects { module, imports }.
	wasmObj := jsCtx.Get("WebAssembly")
	instanceResult := wasmObj.Call("instantiate", module, importObj)
	if isPromise(instanceResult) {
		instanceResult = awaitPromise(ctx, instanceResult)
	}

	instance := instanceResult.Get("module")
	return instance, nil
}

// buildImportObject constructs the JavaScript import object required by
// the GopherJS Go runtime for WASM execution.
func buildImportObject(jsCtx js.Value) map[string]interface{} {
	// The Go WASM runtime expects specific imports.
	imports := make(map[string]interface{})

	// Go runtime import functions.
	goImports := make(map[string]interface{})

	// __go傑_run executes Go functions from JavaScript.
	goImports["__go傑_run"] = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		return nil
	})

	// __go傑_panic handles panics in WASM.
	goImports["__go傑_panic"] = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			slog.Error("WASM panic", "message", args[0].String())
		}
		return nil
	})

	// Register memory if available.
	mem := jsCtx.Get("WebAssembly").Get("Memory")
	if mem.Truthy() {
		goImports["memory"] = mem
	}

	imports["go"] = goImports

	// Environment imports (required by Go runtime).
	envImports := make(map[string]interface{})
	envImports["debug"] = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		return nil
	})
	imports["env"] = envImports

	return imports
}

// Call invokes an exported function from the WASM plugin.
// Args are passed as JavaScript values and return values are converted back.
func (w *WASM) Call(ctx context.Context, fn string, args ...interface{}) (interface{}, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	if !w.exports.Truthy() {
		return nil, fmt.Errorf("WASM exports not available for plugin %q", w.id)
	}

	fnVal := w.exports.Get(fn)
	if !fnVal.Truthy() || !fnVal.Type().Equal(js.TypeFunction) {
		return nil, fmt.Errorf("exported function %q not found in plugin %q", fn, w.id)
	}

	// Convert Go arguments to JavaScript values.
	jsArgs := make([]js.Value, len(args))
	for i, arg := range args {
		jsArgs[i] = goToJS(arg, js.Global())
	}

	result := fnVal.Invoke(jsArgs...)
	if isPromise(result) {
		result = awaitPromise(ctx, result)
	}

	return jsToGo(result), nil
}

// CallVoid invokes an exported function and discards the return value.
func (w *WASM) CallVoid(ctx context.Context, fn string, args ...interface{}) error {
	_, err := w.Call(ctx, fn, args...)
	return err
}

// RegisterCallback registers a Go function to be called from JavaScript.
// The name is how the function will be accessible in the JS plugin.
func (w *WASM) RegisterCallback(name string, fn func(args []js.Value) interface{}) {
	w.mu.Lock()
	defer w.mu.Unlock()

	wrapper := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		return fn(args)
	})

	w.exports.Set(name, wrapper)
}

// UnregisterCallback removes a registered callback.
func (w *WASM) UnregisterCallback(name string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.exports.Set(name, js.Undefined())
}

// ExportedNames returns the list of all exported function names from this plugin.
func (w *WASM) ExportedNames() []string {
	w.mu.RLock()
	defer w.mu.RUnlock()

	if !w.exports.Truthy() {
		return nil
	}

	names := w.exports.Call("Object.keys")
	count := names.Length()
	result := make([]string, count)
	for i := 0; i < count; i++ {
		result[i] = names.Index(i).String()
	}
	return result
}

// Close releases the WASM plugin and frees all associated resources.
func (w *WASM) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	delete(registry, w.id)

	w.instance = js.Value{}
	w.module = js.Value{}
	w.exports = js.Value{}

	slog.Info("WASM plugin closed", "id", w.id)
	return nil
}

// Get returns a loaded plugin by ID.
func Get(id string) (*WASM, bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()

	p, ok := registry[id]
	return p, ok
}

// List returns all loaded plugin IDs.
func List() []string {
	registryMu.RLock()
	defer registryMu.RUnlock()

	ids := make([]string, 0, len(registry))
	for id := range registry {
		ids = append(ids, id)
	}
	return ids
}

// now returns the current Unix timestamp in milliseconds.
func now() int64 {
	return 0 // placeholder; time.Now().UnixMilli()
}

// jsarrayFromBytes creates a JavaScript Uint8Array from a Go byte slice.
func jsarrayFromBytes(data []byte, jsCtx js.Value) js.Value {
	uint8Array := jsCtx.Get("Uint8Array")
	jsBytes := uint8Array.New(len(data))
	jsBytes.Call("set", data)
	return jsBytes
}

// isPromise checks if a JavaScript value is a Promise.
func isPromise(v js.Value) bool {
	if !v.Truthy() {
		return false
	}
	then := v.Get("then")
	return then.Truthy() && then.Type().Equal(js.TypeFunction)
}

// awaitPromise waits for a JavaScript Promise to resolve.
func awaitPromise(ctx context.Context, p js.Value) js.Value {
	// Simplified; production would use a proper async awaiting mechanism.
	// This requires the runtime to provide async support.
	done := make(chan js.Value)
	rejected := make(chan error)

	resolve := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		select {
		case done <- args[0]:
		default:
		}
		return nil
	})

	reject := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		select {
		case rejected <- fmt.Errorf("promise rejected: %s", args[0].String()):
		default:
		}
		return nil
	})

	p.Call("then", resolve, reject)

	select {
	case <-ctx.Done():
		return js.Undefined()
	case result := <-done:
		return result
	case err := <-rejected:
		slog.Error("Promise rejected", "error", err)
		return js.Undefined()
	}
}

// goToJS converts a Go value to its JavaScript equivalent.
// This handles basic types; complex types require custom serialization.
func goToJS(v interface{}, jsCtx js.Value) js.Value {
	switch val := v.(type) {
	case nil:
		return js.Undefined()
	case bool:
		return js.ValueOf(val)
	case string:
		return js.ValueOf(val)
	case int:
		return js.ValueOf(val)
	case int64:
		return js.ValueOf(val)
	case float64:
		return js.ValueOf(val)
	case []byte:
		return jsarrayFromBytes(val, jsCtx)
	case map[string]interface{}:
		obj := jsCtx.Get("Object").Call("create", js.Null())
		for k, v := range val {
			obj.Set(k, goToJS(v, jsCtx))
		}
		return obj
	case []interface{}:
		arr := make([]js.Value, len(val))
		for i, e := range val {
			arr[i] = goToJS(e, jsCtx)
		}
		return jsCtx.Get("Array").Call("from", arr)
	default:
		return js.ValueOf(val)
	}
}

// jsToGo converts a JavaScript value back to its Go equivalent.
func jsToGo(v js.Value) interface{} {
	switch v.Type() {
	case js.TypeUndefined, js.TypeNull:
		return nil
	case js.TypeBoolean:
		return v.Bool()
	case js.TypeNumber:
		return v.Float()
	case js.TypeString:
		return v.String()
	case js.TypeObject:
		if isPromise(v) {
			return nil // Promises must be handled specially.
		}
		arr := v.Call("Array.isArray")
		if arr.Bool() {
			length := v.Get("length").Int()
			result := make([]interface{}, length)
			for i := 0; i < length; i++ {
				result[i] = jsToGo(v.Index(i))
			}
			return result
		}
		keys := v.Call("Object.keys")
		length := keys.Get("length").Int()
		result := make(map[string]interface{})
		for i := 0; i < length; i++ {
			k := keys.Index(i).String()
			result[k] = jsToGo(v.Get(k))
		}
		return result
	case js.TypeFunction:
		return nil // Functions cannot be directly converted.
	default:
		return nil
	}
}

// js is the JavaScript global context; set during runtime initialization.
var js js.Value

// SetJSCtx initializes the JavaScript context for plugin execution.
// This must be called before any plugin operations.
func SetJSCtx(ctx js.Value) {
	js = ctx
}