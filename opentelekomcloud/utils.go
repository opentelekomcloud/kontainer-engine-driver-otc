package opentelekomcloud

import (
	"github.com/rancher/kontainer-engine/drivers/options"
	"github.com/rancher/kontainer-engine/types"
)

var get = options.GetValueFromDriverOptions

type strFromOpts func(keys ...string) string
type strSliceFromOpts func(keys ...string) []string
type intFromOpts func(keys ...string) int64
type boolFromOpts func(keys ...string) bool

// Produce options getters for each argument type
func getters(opts *types.DriverOptions) (strFromOpts, strSliceFromOpts, intFromOpts, boolFromOpts) {
	s := func(k ...string) string {
		return get(opts, types.StringType, k...).(string)
	}
	sl := func(k ...string) []string {
		return get(opts, types.StringSliceType, k...).(*types.StringSlice).Value
	}
	i := func(k ...string) int64 {
		return get(opts, types.IntType, k...).(int64)
	}
	b := func(k ...string) bool {
		return get(opts, types.BoolType, k...).(bool)
	}
	return s, sl, i, b
}
