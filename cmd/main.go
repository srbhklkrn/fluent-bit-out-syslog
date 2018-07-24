package main

import (
	"C"
	"unsafe"

	"log"

	"strings"
	"time"

	"github.com/fluent/fluent-bit-go/output"
	"github.com/pivotal-cf/fluent-bit-out-syslog/pkg/syslog"
)

var out *syslog.Out

//export FLBPluginRegister
func FLBPluginRegister(ctx unsafe.Pointer) int {
	return output.FLBPluginRegister(
		ctx,
		"syslog",
		"syslog output plugin that follows RFC 5424",
	)
}

//export FLBPluginInit
func FLBPluginInit(ctx unsafe.Pointer) int {
	addr := output.FLBPluginConfigKey(ctx, "addr")
	log.Println("[out_syslog] addr = ", addr)

	tls := output.FLBPluginConfigKey(ctx, "enabletls")
	log.Println("[out_syslog] tls = ", tls)

	if strings.EqualFold(tls, "true") {
		skipVerifyS := output.FLBPluginConfigKey(ctx, "insecureskipverify")
		log.Println("[out_syslog] insecure_skip_verify = ", skipVerifyS)

		skipVerify := strings.EqualFold(skipVerifyS, "true")

		out = syslog.NewTLSOut(addr, skipVerify, 30*time.Second)
		return output.FLB_OK
	}

	out = syslog.NewOut(addr)
	return output.FLB_OK
}

//export FLBPluginFlush
func FLBPluginFlush(data unsafe.Pointer, length C.int, tag *C.char) int {
	var (
		ret    int
		ts     interface{}
		record map[interface{}]interface{}
	)

	dec := output.NewDecoder(data, int(length))
	for {
		ret, ts, record = output.GetRecord(dec)
		if ret != 0 {
			break
		}

		flbTime, ok := ts.(output.FLBTime)
		if !ok {
			continue
		}
		timestamp := flbTime.Time

		err := out.Write(record, timestamp, C.GoString(tag))
		if err != nil {
			// TODO: switch over to FLB_RETRY when we are capable of retrying
			// TODO: how we know the flush keeps running issues.
			return output.FLB_ERROR
		}
	}

	return output.FLB_OK
}

//export FLBPluginExit
func FLBPluginExit() int {
	return output.FLB_OK
}

func main() {
}
