package main

import "C"
import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"
	"unsafe"

	"github.com/fluent/fluent-bit-go/input"
	"github.com/go-resty/resty/v2"
)

var rnd = rand.New(rand.NewSource(time.Now().UTC().UnixNano()))

var contexts = make(map[string]*Ctx)

type Ctx struct {
	url    string
	client *resty.Client
}

//export FLBPluginRegister
func FLBPluginRegister(def unsafe.Pointer) int {
	return input.FLBPluginRegister(def, "ghttp", "ghttp GO!")
}

// (fluentbit will call this)
// plugin (context) pointer to fluentbit context (state/ c code)
//
//export FLBPluginInit
func FLBPluginInit(plugin unsafe.Pointer) int {
	return input.FLB_OK
}

//export FLBPluginInputCallback
func FLBPluginInputCallback(data *unsafe.Pointer, size *C.size_t) int {
	now := time.Now()
	flb_time := input.FLBTime{now}
	client := resty.New()
	resp, err := client.R().EnableTrace().Get("http://127.0.0.1:8091/api/v1/instances")
	if err != nil {
		fmt.Println("http get error:", err)
		return input.FLB_ERROR
	}
	bmsg := resp.Body()
	message := make(map[string]interface{}, 0)
	json.Unmarshal(bmsg, &message)
	entry := []interface{}{flb_time, message}

	enc := input.NewEncoder()
	packed, err := enc.Encode(entry)
	if err != nil {
		fmt.Println("Can't convert to msgpack:", message, err)
		return input.FLB_ERROR
	}

	length := len(packed)
	*data = C.CBytes(packed)
	*size = C.size_t(length)
	// For emitting interval adjustment.
	time.Sleep(1000 * time.Millisecond)

	return input.FLB_OK
}

//export FLBPluginInputCleanupCallback
func FLBPluginInputCleanupCallback(data unsafe.Pointer) int {
	return input.FLB_OK
}

//export FLBPluginExit
func FLBPluginExit() int {
	return input.FLB_OK
}

func main() {}
