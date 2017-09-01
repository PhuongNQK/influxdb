package query_test

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/influxdb/influxql"
	"github.com/influxdata/influxdb/query"
)

// Second represents a helper for type converting durations.
const Second = int64(time.Second)

func TestSelect(t *testing.T) {
	for _, tt := range []struct {
		name   string
		q      string
		typ    influxql.DataType
		expr   string
		itrs   []query.Iterator
		points [][]query.Point
		err    string
	}{
		{
			name: "Min",
			q:    `SELECT min(value) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-02T00:00:00Z' GROUP BY time(10s), host fill(none)`,
			typ:  influxql.Float,
			expr: `min(value::float)`,
			itrs: []query.Iterator{
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 0 * Second, Value: 20},
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 11 * Second, Value: 3},
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 31 * Second, Value: 100},
				}},
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 9 * Second, Value: 19},
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 10 * Second, Value: 2},
				}},
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 5 * Second, Value: 10},
				}},
			},
			points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 0 * Second, Value: 19, Aggregated: 2}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 10 * Second, Value: 2, Aggregated: 2}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 30 * Second, Value: 100, Aggregated: 1}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=B"), Time: 0 * Second, Value: 10, Aggregated: 1}},
			},
		},
		{
			name: "Distinct_Float",
			q:    `SELECT distinct(value) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-02T00:00:00Z' GROUP BY time(10s), host fill(none)`,
			typ:  influxql.Float,
			itrs: []query.Iterator{
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 0 * Second, Value: 20},
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 1 * Second, Value: 19},
				}},
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 5 * Second, Value: 10},
				}},
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 9 * Second, Value: 19},
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 10 * Second, Value: 2},
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 11 * Second, Value: 2},
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 12 * Second, Value: 2},
				}},
			},
			points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 0 * Second, Value: 20}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 0 * Second, Value: 19}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 10 * Second, Value: 2}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=B"), Time: 0 * Second, Value: 10}},
			},
		},
		{
			name: "Distinct_Integer",
			q:    `SELECT distinct(value) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-02T00:00:00Z' GROUP BY time(10s), host fill(none)`,
			typ:  influxql.Integer,
			itrs: []query.Iterator{
				&IntegerIterator{Points: []query.IntegerPoint{
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 0 * Second, Value: 20},
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 1 * Second, Value: 19},
				}},
				&IntegerIterator{Points: []query.IntegerPoint{
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 5 * Second, Value: 10},
				}},
				&IntegerIterator{Points: []query.IntegerPoint{
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 9 * Second, Value: 19},
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 10 * Second, Value: 2},
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 11 * Second, Value: 2},
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 12 * Second, Value: 2},
				}},
			},
			points: [][]query.Point{
				{&query.IntegerPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 0 * Second, Value: 20}},
				{&query.IntegerPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 0 * Second, Value: 19}},
				{&query.IntegerPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 10 * Second, Value: 2}},
				{&query.IntegerPoint{Name: "cpu", Tags: ParseTags("host=B"), Time: 0 * Second, Value: 10}},
			},
		},
		{
			name: "Distinct_String",
			q:    `SELECT distinct(value) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-02T00:00:00Z' GROUP BY time(10s), host fill(none)`,
			typ:  influxql.String,
			itrs: []query.Iterator{
				&StringIterator{Points: []query.StringPoint{
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 0 * Second, Value: "a"},
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 1 * Second, Value: "b"},
				}},
				&StringIterator{Points: []query.StringPoint{
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 5 * Second, Value: "c"},
				}},
				&StringIterator{Points: []query.StringPoint{
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 9 * Second, Value: "b"},
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 10 * Second, Value: "d"},
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 11 * Second, Value: "d"},
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 12 * Second, Value: "d"},
				}},
			},
			points: [][]query.Point{
				{&query.StringPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 0 * Second, Value: "a"}},
				{&query.StringPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 0 * Second, Value: "b"}},
				{&query.StringPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 10 * Second, Value: "d"}},
				{&query.StringPoint{Name: "cpu", Tags: ParseTags("host=B"), Time: 0 * Second, Value: "c"}},
			},
		},
		{
			name: "Distinct_Boolean",
			q:    `SELECT distinct(value) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-02T00:00:00Z' GROUP BY time(10s), host fill(none)`,
			typ:  influxql.Boolean,
			itrs: []query.Iterator{
				&BooleanIterator{Points: []query.BooleanPoint{
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 0 * Second, Value: true},
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 1 * Second, Value: false},
				}},
				&BooleanIterator{Points: []query.BooleanPoint{
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 5 * Second, Value: false},
				}},
				&BooleanIterator{Points: []query.BooleanPoint{
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 9 * Second, Value: true},
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 10 * Second, Value: false},
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 11 * Second, Value: false},
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 12 * Second, Value: true},
				}},
			},
			points: [][]query.Point{
				{&query.BooleanPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 0 * Second, Value: true}},
				{&query.BooleanPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 0 * Second, Value: false}},
				{&query.BooleanPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 10 * Second, Value: false}},
				{&query.BooleanPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 10 * Second, Value: true}},
				{&query.BooleanPoint{Name: "cpu", Tags: ParseTags("host=B"), Time: 0 * Second, Value: false}},
			},
		},
		{
			name: "Mean_Float",
			q:    `SELECT mean(value) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-02T00:00:00Z' GROUP BY time(10s), host fill(none)`,
			typ:  influxql.Float,
			expr: `mean(value::float)`,
			itrs: []query.Iterator{
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 0 * Second, Value: 20},
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 11 * Second, Value: 3},
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 31 * Second, Value: 100},
				}},
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 5 * Second, Value: 10},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 50 * Second, Value: 1},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 51 * Second, Value: 2},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 52 * Second, Value: 4},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 53 * Second, Value: 4},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 53 * Second, Value: 5},
				}},
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 9 * Second, Value: 19},
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 10 * Second, Value: 2},
				}},
			},
			points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 0 * Second, Value: 19.5, Aggregated: 2}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 10 * Second, Value: 2.5, Aggregated: 2}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 30 * Second, Value: 100, Aggregated: 1}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=B"), Time: 0 * Second, Value: 10, Aggregated: 1}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=B"), Time: 50 * Second, Value: 3.2, Aggregated: 5}},
			},
		},
		{
			name: "Mean_Integer",
			q:    `SELECT mean(value) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-02T00:00:00Z' GROUP BY time(10s), host fill(none)`,
			typ:  influxql.Integer,
			expr: `mean(value::integer)`,
			itrs: []query.Iterator{
				&IntegerIterator{Points: []query.IntegerPoint{
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 0 * Second, Value: 20},
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 11 * Second, Value: 3},
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 31 * Second, Value: 100},
				}},
				&IntegerIterator{Points: []query.IntegerPoint{
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 5 * Second, Value: 10},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 50 * Second, Value: 1},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 51 * Second, Value: 2},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 52 * Second, Value: 4},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 53 * Second, Value: 4},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 53 * Second, Value: 5},
				}},
				&IntegerIterator{Points: []query.IntegerPoint{
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 9 * Second, Value: 19},
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 10 * Second, Value: 2},
				}},
			},
			points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 0 * Second, Value: 19.5, Aggregated: 2}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 10 * Second, Value: 2.5, Aggregated: 2}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 30 * Second, Value: 100, Aggregated: 1}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=B"), Time: 0 * Second, Value: 10, Aggregated: 1}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=B"), Time: 50 * Second, Value: 3.2, Aggregated: 5}},
			},
		},
		{
			name: "Mean_String",
			q:    `SELECT mean(value) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-02T00:00:00Z' GROUP BY time(10s), host fill(none)`,
			typ:  influxql.String,
			itrs: []query.Iterator{&StringIterator{}},
			err:  `unsupported mean iterator type: *query_test.StringIterator`,
		},
		{
			name: "Mean_Boolean",
			q:    `SELECT mean(value) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-02T00:00:00Z' GROUP BY time(10s), host fill(none)`,
			typ:  influxql.Boolean,
			itrs: []query.Iterator{&BooleanIterator{}},
			err:  `unsupported mean iterator type: *query_test.BooleanIterator`,
		},
		{
			name: "Median_Float",
			q:    `SELECT median(value) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-02T00:00:00Z' GROUP BY time(10s), host fill(none)`,
			typ:  influxql.Float,
			itrs: []query.Iterator{
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 0 * Second, Value: 20},
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 11 * Second, Value: 3},
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 31 * Second, Value: 100},
				}},
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 5 * Second, Value: 10},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 50 * Second, Value: 1},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 51 * Second, Value: 2},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 52 * Second, Value: 3},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 53 * Second, Value: 4},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 53 * Second, Value: 5},
				}},
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 9 * Second, Value: 19},
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 10 * Second, Value: 2},
				}},
			},
			points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 0 * Second, Value: 19.5}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 10 * Second, Value: 2.5}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 30 * Second, Value: 100}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=B"), Time: 0 * Second, Value: 10}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=B"), Time: 50 * Second, Value: 3}},
			},
		},
		{
			name: "Median_Integer",
			q:    `SELECT median(value) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-02T00:00:00Z' GROUP BY time(10s), host fill(none)`,
			typ:  influxql.Integer,
			itrs: []query.Iterator{
				&IntegerIterator{Points: []query.IntegerPoint{
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 0 * Second, Value: 20},
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 11 * Second, Value: 3},
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 31 * Second, Value: 100},
				}},
				&IntegerIterator{Points: []query.IntegerPoint{
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 5 * Second, Value: 10},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 50 * Second, Value: 1},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 51 * Second, Value: 2},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 52 * Second, Value: 3},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 53 * Second, Value: 4},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 53 * Second, Value: 5},
				}},
				&IntegerIterator{Points: []query.IntegerPoint{
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 9 * Second, Value: 19},
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 10 * Second, Value: 2},
				}},
			},
			points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 0 * Second, Value: 19.5}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 10 * Second, Value: 2.5}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 30 * Second, Value: 100}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=B"), Time: 0 * Second, Value: 10}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=B"), Time: 50 * Second, Value: 3}},
			},
		},
		{
			name: "Median_String",
			q:    `SELECT median(value) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-02T00:00:00Z' GROUP BY time(10s), host fill(none)`,
			typ:  influxql.String,
			itrs: []query.Iterator{&StringIterator{}},
			err:  `unsupported median iterator type: *query_test.StringIterator`,
		},
		{
			name: "Median_Boolean",
			q:    `SELECT median(value) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-02T00:00:00Z' GROUP BY time(10s), host fill(none)`,
			typ:  influxql.Boolean,
			itrs: []query.Iterator{&BooleanIterator{}},
			err:  `unsupported median iterator type: *query_test.BooleanIterator`,
		},
		{
			name: "Mode_Float",
			q:    `SELECT mode(value) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-02T00:00:00Z' GROUP BY time(10s), host fill(none)`,
			typ:  influxql.Float,
			itrs: []query.Iterator{
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 0 * Second, Value: 10},
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 11 * Second, Value: 2},
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 31 * Second, Value: 100},
				}},
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 5 * Second, Value: 10},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 50 * Second, Value: 1},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 51 * Second, Value: 2},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 52 * Second, Value: 3},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 53 * Second, Value: 4},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 53 * Second, Value: 5},
				}},
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 9 * Second, Value: 19},
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 10 * Second, Value: 2},
				}},
			},
			points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 0 * Second, Value: 10}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 10 * Second, Value: 2}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 30 * Second, Value: 100}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=B"), Time: 0 * Second, Value: 10}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=B"), Time: 50 * Second, Value: 1}},
			},
		},
		{
			name: "Mode_Integer",
			q:    `SELECT mode(value) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-02T00:00:00Z' GROUP BY time(10s), host fill(none)`,
			typ:  influxql.Integer,
			itrs: []query.Iterator{
				&IntegerIterator{Points: []query.IntegerPoint{
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 0 * Second, Value: 10},
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 11 * Second, Value: 2},
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 31 * Second, Value: 100},
				}},
				&IntegerIterator{Points: []query.IntegerPoint{
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 5 * Second, Value: 10},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 50 * Second, Value: 1},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 51 * Second, Value: 2},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 52 * Second, Value: 3},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 53 * Second, Value: 4},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 54 * Second, Value: 5},
				}},
				&IntegerIterator{Points: []query.IntegerPoint{
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 9 * Second, Value: 19},
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 10 * Second, Value: 2},
				}},
			},
			points: [][]query.Point{
				{&query.IntegerPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 0 * Second, Value: 10}},
				{&query.IntegerPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 10 * Second, Value: 2}},
				{&query.IntegerPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 30 * Second, Value: 100}},
				{&query.IntegerPoint{Name: "cpu", Tags: ParseTags("host=B"), Time: 0 * Second, Value: 10}},
				{&query.IntegerPoint{Name: "cpu", Tags: ParseTags("host=B"), Time: 50 * Second, Value: 1}},
			},
		},
		{
			name: "Mode_String",
			q:    `SELECT mode(value) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-02T00:00:00Z' GROUP BY time(10s), host fill(none)`,
			typ:  influxql.String,
			itrs: []query.Iterator{
				&StringIterator{Points: []query.StringPoint{
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 0 * Second, Value: "a"},
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 1 * Second, Value: "a"},
				}},
				&StringIterator{Points: []query.StringPoint{
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 5 * Second, Value: "cxxx"},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 6 * Second, Value: "zzzz"},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 7 * Second, Value: "zzzz"},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 8 * Second, Value: "zxxx"},
				}},
				&StringIterator{Points: []query.StringPoint{
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 9 * Second, Value: "b"},
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 10 * Second, Value: "d"},
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 11 * Second, Value: "d"},
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 12 * Second, Value: "d"},
				}},
			},
			points: [][]query.Point{
				{&query.StringPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 0 * Second, Value: "a"}},
				{&query.StringPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 10 * Second, Value: "d"}},
				{&query.StringPoint{Name: "cpu", Tags: ParseTags("host=B"), Time: 0 * Second, Value: "zzzz"}},
			},
		},
		{
			name: "Mode_Boolean",
			q:    `SELECT mode(value) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-02T00:00:00Z' GROUP BY time(10s), host fill(none)`,
			typ:  influxql.Boolean,
			itrs: []query.Iterator{
				&BooleanIterator{Points: []query.BooleanPoint{
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 0 * Second, Value: true},
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 1 * Second, Value: false},
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 2 * Second, Value: false},
				}},
				&BooleanIterator{Points: []query.BooleanPoint{
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 5 * Second, Value: true},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 6 * Second, Value: false},
				}},
				&BooleanIterator{Points: []query.BooleanPoint{
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 9 * Second, Value: false},
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 10 * Second, Value: true},
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 11 * Second, Value: false},
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 12 * Second, Value: true},
				}},
			},
			points: [][]query.Point{
				{&query.BooleanPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 0 * Second, Value: false}},
				{&query.BooleanPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 10 * Second, Value: true}},
				{&query.BooleanPoint{Name: "cpu", Tags: ParseTags("host=B"), Time: 0 * Second, Value: true}},
			},
		},
		{
			name: "Top_NoTags_Float",
			q:    `SELECT top(value, 2) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-02T00:00:00Z' GROUP BY time(30s), host fill(none)`,
			typ:  influxql.Float,
			itrs: []query.Iterator{
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 0 * Second, Value: 20},
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 11 * Second, Value: 3},
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 31 * Second, Value: 100},
				}},
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 5 * Second, Value: 10},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 50 * Second, Value: 1},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 51 * Second, Value: 2},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 52 * Second, Value: 3},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 53 * Second, Value: 4},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 53 * Second, Value: 5},
				}},
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 9 * Second, Value: 19},
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 10 * Second, Value: 2},
				}},
			},
			points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 0 * Second, Value: 20}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 9 * Second, Value: 19}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 31 * Second, Value: 100}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=B"), Time: 5 * Second, Value: 10}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=B"), Time: 53 * Second, Value: 5}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=B"), Time: 53 * Second, Value: 4}},
			},
		},
		{
			name: "Top_NoTags_Integer",
			q:    `SELECT top(value, 2) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-02T00:00:00Z' GROUP BY time(30s), host fill(none)`,
			typ:  influxql.Integer,
			itrs: []query.Iterator{
				&IntegerIterator{Points: []query.IntegerPoint{
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 0 * Second, Value: 20},
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 11 * Second, Value: 3},
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 31 * Second, Value: 100},
				}},
				&IntegerIterator{Points: []query.IntegerPoint{
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 5 * Second, Value: 10},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 50 * Second, Value: 1},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 51 * Second, Value: 2},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 52 * Second, Value: 3},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 53 * Second, Value: 4},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 53 * Second, Value: 5},
				}},
				&IntegerIterator{Points: []query.IntegerPoint{
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 9 * Second, Value: 19},
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 10 * Second, Value: 2},
				}},
			},
			points: [][]query.Point{
				{&query.IntegerPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 0 * Second, Value: 20}},
				{&query.IntegerPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 9 * Second, Value: 19}},
				{&query.IntegerPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 31 * Second, Value: 100}},
				{&query.IntegerPoint{Name: "cpu", Tags: ParseTags("host=B"), Time: 5 * Second, Value: 10}},
				{&query.IntegerPoint{Name: "cpu", Tags: ParseTags("host=B"), Time: 53 * Second, Value: 5}},
				{&query.IntegerPoint{Name: "cpu", Tags: ParseTags("host=B"), Time: 53 * Second, Value: 4}},
			},
		},
		{
			name: "Top_Tags_Float",
			q:    `SELECT top(value::float, host::tag, 2) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-02T00:00:00Z' GROUP BY time(30s) fill(none)`,
			typ:  influxql.Float,
			expr: `max(value::float)`,
			itrs: []query.Iterator{
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 0 * Second, Value: 20, Aux: []interface{}{"A"}},
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 11 * Second, Value: 3, Aux: []interface{}{"A"}},
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 31 * Second, Value: 100, Aux: []interface{}{"A"}},
				}},
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 5 * Second, Value: 10, Aux: []interface{}{"B"}},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 50 * Second, Value: 1, Aux: []interface{}{"B"}},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 51 * Second, Value: 2, Aux: []interface{}{"B"}},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 52 * Second, Value: 3, Aux: []interface{}{"B"}},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 53 * Second, Value: 4, Aux: []interface{}{"B"}},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 53 * Second, Value: 5, Aux: []interface{}{"B"}},
				}},
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 9 * Second, Value: 19, Aux: []interface{}{"A"}},
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 10 * Second, Value: 2, Aux: []interface{}{"A"}},
				}},
			},
			points: [][]query.Point{
				{
					&query.FloatPoint{Name: "cpu", Time: 0 * Second, Value: 20, Aux: []interface{}{"A"}},
					&query.StringPoint{Name: "cpu", Time: 0 * Second, Value: "A"},
				},
				{
					&query.FloatPoint{Name: "cpu", Time: 5 * Second, Value: 10, Aux: []interface{}{"B"}},
					&query.StringPoint{Name: "cpu", Time: 5 * Second, Value: "B"},
				},
				{
					&query.FloatPoint{Name: "cpu", Time: 31 * Second, Value: 100, Aux: []interface{}{"A"}},
					&query.StringPoint{Name: "cpu", Time: 31 * Second, Value: "A"},
				},
				{
					&query.FloatPoint{Name: "cpu", Time: 53 * Second, Value: 5, Aux: []interface{}{"B"}},
					&query.StringPoint{Name: "cpu", Time: 53 * Second, Value: "B"},
				},
			},
		},
		{
			name: "Top_Tags_Integer",
			q:    `SELECT top(value::integer, host::tag, 2) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-02T00:00:00Z' GROUP BY time(30s) fill(none)`,
			typ:  influxql.Integer,
			expr: `max(value::integer)`,
			itrs: []query.Iterator{
				&IntegerIterator{Points: []query.IntegerPoint{
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 0 * Second, Value: 20, Aux: []interface{}{"A"}},
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 11 * Second, Value: 3, Aux: []interface{}{"A"}},
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 31 * Second, Value: 100, Aux: []interface{}{"A"}},
				}},
				&IntegerIterator{Points: []query.IntegerPoint{
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 5 * Second, Value: 10, Aux: []interface{}{"B"}},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 50 * Second, Value: 1, Aux: []interface{}{"B"}},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 51 * Second, Value: 2, Aux: []interface{}{"B"}},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 52 * Second, Value: 3, Aux: []interface{}{"B"}},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 53 * Second, Value: 4, Aux: []interface{}{"B"}},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 53 * Second, Value: 5, Aux: []interface{}{"B"}},
				}},
				&IntegerIterator{Points: []query.IntegerPoint{
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 9 * Second, Value: 19, Aux: []interface{}{"A"}},
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 10 * Second, Value: 2, Aux: []interface{}{"A"}},
				}},
			},
			points: [][]query.Point{
				{
					&query.IntegerPoint{Name: "cpu", Time: 0 * Second, Value: 20, Aux: []interface{}{"A"}},
					&query.StringPoint{Name: "cpu", Time: 0 * Second, Value: "A"},
				},
				{
					&query.IntegerPoint{Name: "cpu", Time: 5 * Second, Value: 10, Aux: []interface{}{"B"}},
					&query.StringPoint{Name: "cpu", Time: 5 * Second, Value: "B"},
				},
				{
					&query.IntegerPoint{Name: "cpu", Time: 31 * Second, Value: 100, Aux: []interface{}{"A"}},
					&query.StringPoint{Name: "cpu", Time: 31 * Second, Value: "A"},
				},
				{
					&query.IntegerPoint{Name: "cpu", Time: 53 * Second, Value: 5, Aux: []interface{}{"B"}},
					&query.StringPoint{Name: "cpu", Time: 53 * Second, Value: "B"},
				},
			},
		},
		{
			name: "Top_GroupByTags_Float",
			q:    `SELECT top(value::float, host::tag, 1) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-02T00:00:00Z' GROUP BY region, time(30s) fill(none)`,
			typ:  influxql.Float,
			expr: `max(value::float)`,
			itrs: []query.Iterator{
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 0 * Second, Value: 20, Aux: []interface{}{"A"}},
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 11 * Second, Value: 3, Aux: []interface{}{"A"}},
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 31 * Second, Value: 100, Aux: []interface{}{"A"}},
				}},
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 5 * Second, Value: 10, Aux: []interface{}{"B"}},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 50 * Second, Value: 1, Aux: []interface{}{"B"}},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 51 * Second, Value: 2, Aux: []interface{}{"B"}},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 52 * Second, Value: 3, Aux: []interface{}{"B"}},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 53 * Second, Value: 4, Aux: []interface{}{"B"}},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 53 * Second, Value: 5, Aux: []interface{}{"B"}},
				}},
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 9 * Second, Value: 19, Aux: []interface{}{"A"}},
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 10 * Second, Value: 2, Aux: []interface{}{"A"}},
				}},
			},
			points: [][]query.Point{
				{
					&query.FloatPoint{Name: "cpu", Tags: ParseTags("region=east"), Time: 9 * Second, Value: 19, Aux: []interface{}{"A"}},
					&query.StringPoint{Name: "cpu", Tags: ParseTags("region=east"), Time: 9 * Second, Value: "A"},
				},
				{
					&query.FloatPoint{Name: "cpu", Tags: ParseTags("region=west"), Time: 0 * Second, Value: 20, Aux: []interface{}{"A"}},
					&query.StringPoint{Name: "cpu", Tags: ParseTags("region=west"), Time: 0 * Second, Value: "A"},
				},
				{
					&query.FloatPoint{Name: "cpu", Tags: ParseTags("region=west"), Time: 31 * Second, Value: 100, Aux: []interface{}{"A"}},
					&query.StringPoint{Name: "cpu", Tags: ParseTags("region=west"), Time: 31 * Second, Value: "A"},
				},
			},
		},
		{
			name: "Top_GroupByTags_Integer",
			q:    `SELECT top(value::integer, host::tag, 1) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-02T00:00:00Z' GROUP BY region, time(30s) fill(none)`,
			typ:  influxql.Integer,
			itrs: []query.Iterator{
				&IntegerIterator{Points: []query.IntegerPoint{
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 0 * Second, Value: 20, Aux: []interface{}{"A"}},
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 11 * Second, Value: 3, Aux: []interface{}{"A"}},
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 31 * Second, Value: 100, Aux: []interface{}{"A"}},
				}},
				&IntegerIterator{Points: []query.IntegerPoint{
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 5 * Second, Value: 10, Aux: []interface{}{"B"}},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 50 * Second, Value: 1, Aux: []interface{}{"B"}},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 51 * Second, Value: 2, Aux: []interface{}{"B"}},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 52 * Second, Value: 3, Aux: []interface{}{"B"}},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 53 * Second, Value: 4, Aux: []interface{}{"B"}},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 53 * Second, Value: 5, Aux: []interface{}{"B"}},
				}},
				&IntegerIterator{Points: []query.IntegerPoint{
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 9 * Second, Value: 19, Aux: []interface{}{"A"}},
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 10 * Second, Value: 2, Aux: []interface{}{"A"}},
				}},
			},
			points: [][]query.Point{
				{
					&query.IntegerPoint{Name: "cpu", Tags: ParseTags("region=east"), Time: 9 * Second, Value: 19, Aux: []interface{}{"A"}},
					&query.StringPoint{Name: "cpu", Tags: ParseTags("region=east"), Time: 9 * Second, Value: "A"},
				},
				{
					&query.IntegerPoint{Name: "cpu", Tags: ParseTags("region=west"), Time: 0 * Second, Value: 20, Aux: []interface{}{"A"}},
					&query.StringPoint{Name: "cpu", Tags: ParseTags("region=west"), Time: 0 * Second, Value: "A"},
				},
				{
					&query.IntegerPoint{Name: "cpu", Tags: ParseTags("region=west"), Time: 31 * Second, Value: 100, Aux: []interface{}{"A"}},
					&query.StringPoint{Name: "cpu", Tags: ParseTags("region=west"), Time: 31 * Second, Value: "A"},
				},
			},
		},
		{
			name: "Bottom_NoTags_Float",
			q:    `SELECT bottom(value::float, 2) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-02T00:00:00Z' GROUP BY time(30s), host fill(none)`,
			typ:  influxql.Float,
			itrs: []query.Iterator{
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 0 * Second, Value: 20},
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 11 * Second, Value: 3},
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 31 * Second, Value: 100},
				}},
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 5 * Second, Value: 10},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 50 * Second, Value: 1},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 51 * Second, Value: 2},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 52 * Second, Value: 3},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 53 * Second, Value: 4},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 53 * Second, Value: 5},
				}},
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 9 * Second, Value: 19},
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 10 * Second, Value: 2},
				}},
			},
			points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 10 * Second, Value: 2}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 11 * Second, Value: 3}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 31 * Second, Value: 100}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=B"), Time: 5 * Second, Value: 10}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=B"), Time: 50 * Second, Value: 1}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=B"), Time: 51 * Second, Value: 2}},
			},
		},
		{
			name: "Bottom_NoTags_Integer",
			q:    `SELECT bottom(value::integer, 2) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-02T00:00:00Z' GROUP BY time(30s), host fill(none)`,
			typ:  influxql.Integer,
			itrs: []query.Iterator{
				&IntegerIterator{Points: []query.IntegerPoint{
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 0 * Second, Value: 20},
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 11 * Second, Value: 3},
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 31 * Second, Value: 100},
				}},
				&IntegerIterator{Points: []query.IntegerPoint{
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 5 * Second, Value: 10},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 50 * Second, Value: 1},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 51 * Second, Value: 2},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 52 * Second, Value: 3},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 53 * Second, Value: 4},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 53 * Second, Value: 5},
				}},
				&IntegerIterator{Points: []query.IntegerPoint{
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 9 * Second, Value: 19},
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 10 * Second, Value: 2},
				}},
			},
			points: [][]query.Point{
				{&query.IntegerPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 10 * Second, Value: 2}},
				{&query.IntegerPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 11 * Second, Value: 3}},
				{&query.IntegerPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 31 * Second, Value: 100}},
				{&query.IntegerPoint{Name: "cpu", Tags: ParseTags("host=B"), Time: 5 * Second, Value: 10}},
				{&query.IntegerPoint{Name: "cpu", Tags: ParseTags("host=B"), Time: 50 * Second, Value: 1}},
				{&query.IntegerPoint{Name: "cpu", Tags: ParseTags("host=B"), Time: 51 * Second, Value: 2}},
			},
		},
		{
			name: "Bottom_Tags_Float",
			q:    `SELECT bottom(value::float, host::tag, 2) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-02T00:00:00Z' GROUP BY time(30s) fill(none)`,
			typ:  influxql.Float,
			expr: `min(value::float)`,
			itrs: []query.Iterator{
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 0 * Second, Value: 20, Aux: []interface{}{"A"}},
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 11 * Second, Value: 3, Aux: []interface{}{"A"}},
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 31 * Second, Value: 100, Aux: []interface{}{"A"}},
				}},
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 5 * Second, Value: 10, Aux: []interface{}{"B"}},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 50 * Second, Value: 1, Aux: []interface{}{"B"}},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 51 * Second, Value: 2, Aux: []interface{}{"B"}},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 52 * Second, Value: 3, Aux: []interface{}{"B"}},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 53 * Second, Value: 4, Aux: []interface{}{"B"}},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 53 * Second, Value: 5, Aux: []interface{}{"B"}},
				}},
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 9 * Second, Value: 19, Aux: []interface{}{"A"}},
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 10 * Second, Value: 2, Aux: []interface{}{"A"}},
				}},
			},
			points: [][]query.Point{
				{
					&query.FloatPoint{Name: "cpu", Time: 5 * Second, Value: 10, Aux: []interface{}{"B"}},
					&query.StringPoint{Name: "cpu", Time: 5 * Second, Value: "B"},
				},
				{
					&query.FloatPoint{Name: "cpu", Time: 10 * Second, Value: 2, Aux: []interface{}{"A"}},
					&query.StringPoint{Name: "cpu", Time: 10 * Second, Value: "A"},
				},
				{
					&query.FloatPoint{Name: "cpu", Time: 31 * Second, Value: 100, Aux: []interface{}{"A"}},
					&query.StringPoint{Name: "cpu", Time: 31 * Second, Value: "A"},
				},
				{
					&query.FloatPoint{Name: "cpu", Time: 50 * Second, Value: 1, Aux: []interface{}{"B"}},
					&query.StringPoint{Name: "cpu", Time: 50 * Second, Value: "B"},
				},
			},
		},
		{
			name: "Bottom_Tags_Integer",
			q:    `SELECT bottom(value::integer, host::tag, 2) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-02T00:00:00Z' GROUP BY time(30s) fill(none)`,
			typ:  influxql.Integer,
			itrs: []query.Iterator{
				&IntegerIterator{Points: []query.IntegerPoint{
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 0 * Second, Value: 20, Aux: []interface{}{"A"}},
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 11 * Second, Value: 3, Aux: []interface{}{"A"}},
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 31 * Second, Value: 100, Aux: []interface{}{"A"}},
				}},
				&IntegerIterator{Points: []query.IntegerPoint{
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 5 * Second, Value: 10, Aux: []interface{}{"B"}},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 50 * Second, Value: 1, Aux: []interface{}{"B"}},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 51 * Second, Value: 2, Aux: []interface{}{"B"}},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 52 * Second, Value: 3, Aux: []interface{}{"B"}},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 53 * Second, Value: 4, Aux: []interface{}{"B"}},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 53 * Second, Value: 5, Aux: []interface{}{"B"}},
				}},
				&IntegerIterator{Points: []query.IntegerPoint{
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 9 * Second, Value: 19, Aux: []interface{}{"A"}},
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 10 * Second, Value: 2, Aux: []interface{}{"A"}},
				}},
			},
			points: [][]query.Point{
				{
					&query.IntegerPoint{Name: "cpu", Time: 5 * Second, Value: 10, Aux: []interface{}{"B"}},
					&query.StringPoint{Name: "cpu", Time: 5 * Second, Value: "B"},
				},
				{
					&query.IntegerPoint{Name: "cpu", Time: 10 * Second, Value: 2, Aux: []interface{}{"A"}},
					&query.StringPoint{Name: "cpu", Time: 10 * Second, Value: "A"},
				},
				{
					&query.IntegerPoint{Name: "cpu", Time: 31 * Second, Value: 100, Aux: []interface{}{"A"}},
					&query.StringPoint{Name: "cpu", Time: 31 * Second, Value: "A"},
				},
				{
					&query.IntegerPoint{Name: "cpu", Time: 50 * Second, Value: 1, Aux: []interface{}{"B"}},
					&query.StringPoint{Name: "cpu", Time: 50 * Second, Value: "B"},
				},
			},
		},
		{
			name: "Bottom_GroupByTags_Float",
			q:    `SELECT bottom(value::float, host::tag, 1) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-02T00:00:00Z' GROUP BY region, time(30s) fill(none)`,
			typ:  influxql.Float,
			expr: `min(value::float)`,
			itrs: []query.Iterator{
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 0 * Second, Value: 20, Aux: []interface{}{"A"}},
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 11 * Second, Value: 3, Aux: []interface{}{"A"}},
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 31 * Second, Value: 100, Aux: []interface{}{"A"}},
				}},
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 5 * Second, Value: 10, Aux: []interface{}{"B"}},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 50 * Second, Value: 1, Aux: []interface{}{"B"}},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 51 * Second, Value: 2, Aux: []interface{}{"B"}},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 52 * Second, Value: 3, Aux: []interface{}{"B"}},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 53 * Second, Value: 4, Aux: []interface{}{"B"}},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 53 * Second, Value: 5, Aux: []interface{}{"B"}},
				}},
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 9 * Second, Value: 19, Aux: []interface{}{"A"}},
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 10 * Second, Value: 2, Aux: []interface{}{"A"}},
				}},
			},
			points: [][]query.Point{
				{
					&query.FloatPoint{Name: "cpu", Tags: ParseTags("region=east"), Time: 10 * Second, Value: 2, Aux: []interface{}{"A"}},
					&query.StringPoint{Name: "cpu", Tags: ParseTags("region=east"), Time: 10 * Second, Value: "A"},
				},
				{
					&query.FloatPoint{Name: "cpu", Tags: ParseTags("region=west"), Time: 11 * Second, Value: 3, Aux: []interface{}{"A"}},
					&query.StringPoint{Name: "cpu", Tags: ParseTags("region=west"), Time: 11 * Second, Value: "A"},
				},
				{
					&query.FloatPoint{Name: "cpu", Tags: ParseTags("region=west"), Time: 50 * Second, Value: 1, Aux: []interface{}{"B"}},
					&query.StringPoint{Name: "cpu", Tags: ParseTags("region=west"), Time: 50 * Second, Value: "B"},
				},
			},
		},
		{
			name: "Bottom_GroupByTags_Integer",
			q:    `SELECT bottom(value::float, host::tag, 1) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-02T00:00:00Z' GROUP BY region, time(30s) fill(none)`,
			typ:  influxql.Integer,
			expr: `min(value::float)`,
			itrs: []query.Iterator{
				&IntegerIterator{Points: []query.IntegerPoint{
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 0 * Second, Value: 20, Aux: []interface{}{"A"}},
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 11 * Second, Value: 3, Aux: []interface{}{"A"}},
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 31 * Second, Value: 100, Aux: []interface{}{"A"}},
				}},
				&IntegerIterator{Points: []query.IntegerPoint{
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 5 * Second, Value: 10, Aux: []interface{}{"B"}},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 50 * Second, Value: 1, Aux: []interface{}{"B"}},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 51 * Second, Value: 2, Aux: []interface{}{"B"}},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 52 * Second, Value: 3, Aux: []interface{}{"B"}},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 53 * Second, Value: 4, Aux: []interface{}{"B"}},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 53 * Second, Value: 5, Aux: []interface{}{"B"}},
				}},
				&IntegerIterator{Points: []query.IntegerPoint{
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 9 * Second, Value: 19, Aux: []interface{}{"A"}},
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 10 * Second, Value: 2, Aux: []interface{}{"A"}},
				}},
			},
			points: [][]query.Point{
				{
					&query.IntegerPoint{Name: "cpu", Tags: ParseTags("region=east"), Time: 10 * Second, Value: 2, Aux: []interface{}{"A"}},
					&query.StringPoint{Name: "cpu", Tags: ParseTags("region=east"), Time: 10 * Second, Value: "A"},
				},
				{
					&query.IntegerPoint{Name: "cpu", Tags: ParseTags("region=west"), Time: 11 * Second, Value: 3, Aux: []interface{}{"A"}},
					&query.StringPoint{Name: "cpu", Tags: ParseTags("region=west"), Time: 11 * Second, Value: "A"},
				},
				{
					&query.IntegerPoint{Name: "cpu", Tags: ParseTags("region=west"), Time: 50 * Second, Value: 1, Aux: []interface{}{"B"}},
					&query.StringPoint{Name: "cpu", Tags: ParseTags("region=west"), Time: 50 * Second, Value: "B"},
				},
			},
		},
		{
			name: "Fill_Null_Float",
			q:    `SELECT mean(value) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-01T00:01:00Z' GROUP BY host, time(10s) fill(null)`,
			typ:  influxql.Float,
			expr: `mean(value::float)`,
			itrs: []query.Iterator{
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Tags: ParseTags("host=A"), Time: 12 * Second, Value: 2},
				}},
			},
			points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 0 * Second, Nil: true}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 10 * Second, Value: 2, Aggregated: 1}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 20 * Second, Nil: true}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 30 * Second, Nil: true}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 40 * Second, Nil: true}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 50 * Second, Nil: true}},
			},
		},
		{
			name: "Fill_Number_Float",
			q:    `SELECT mean(value) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-01T00:01:00Z' GROUP BY host, time(10s) fill(1)`,
			typ:  influxql.Float,
			expr: `mean(value::float)`,
			itrs: []query.Iterator{
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Tags: ParseTags("host=A"), Time: 12 * Second, Value: 2},
				}},
			},
			points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 0 * Second, Value: 1}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 10 * Second, Value: 2, Aggregated: 1}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 20 * Second, Value: 1}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 30 * Second, Value: 1}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 40 * Second, Value: 1}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 50 * Second, Value: 1}},
			},
		},
		{
			name: "Fill_Previous_Float",
			q:    `SELECT mean(value) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-01T00:01:00Z' GROUP BY host, time(10s) fill(previous)`,
			typ:  influxql.Float,
			expr: `mean(value::float)`,
			itrs: []query.Iterator{
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Tags: ParseTags("host=A"), Time: 12 * Second, Value: 2},
				}},
			},
			points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 0 * Second, Nil: true}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 10 * Second, Value: 2, Aggregated: 1}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 20 * Second, Value: 2}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 30 * Second, Value: 2}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 40 * Second, Value: 2}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 50 * Second, Value: 2}},
			},
		},
		{
			name: "Fill_Linear_Float_One",
			q:    `SELECT mean(value) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-01T00:01:00Z' GROUP BY host, time(10s) fill(linear)`,
			typ:  influxql.Float,
			expr: `mean(value::float)`,
			itrs: []query.Iterator{
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Tags: ParseTags("host=A"), Time: 12 * Second, Value: 2},
					{Name: "cpu", Tags: ParseTags("host=A"), Time: 32 * Second, Value: 4},
				}},
			},
			points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 0 * Second, Nil: true}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 10 * Second, Value: 2, Aggregated: 1}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 20 * Second, Value: 3}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 30 * Second, Value: 4, Aggregated: 1}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 40 * Second, Nil: true}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 50 * Second, Nil: true}},
			},
		},
		{
			name: "Fill_Linear_Float_Many",
			q:    `SELECT mean(value) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-01T00:01:00Z' GROUP BY host, time(10s) fill(linear)`,
			typ:  influxql.Float,
			expr: `mean(value::float)`,
			itrs: []query.Iterator{
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Tags: ParseTags("host=A"), Time: 12 * Second, Value: 2},
					{Name: "cpu", Tags: ParseTags("host=A"), Time: 62 * Second, Value: 7},
				}},
			},
			points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 0 * Second, Nil: true}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 10 * Second, Value: 2, Aggregated: 1}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 20 * Second, Value: 3}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 30 * Second, Value: 4}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 40 * Second, Value: 5}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 50 * Second, Value: 6}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 60 * Second, Value: 7, Aggregated: 1}},
			},
		},
		{
			name: "Fill_Linear_Float_MultipleSeries",
			q:    `SELECT mean(value) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-01T00:01:00Z' GROUP BY host, time(10s) fill(linear)`,
			typ:  influxql.Float,
			expr: `mean(value::float)`,
			itrs: []query.Iterator{
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Tags: ParseTags("host=A"), Time: 12 * Second, Value: 2},
					{Name: "cpu", Tags: ParseTags("host=B"), Time: 32 * Second, Value: 4},
				}},
			},
			points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 0 * Second, Nil: true}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 10 * Second, Value: 2, Aggregated: 1}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 20 * Second, Nil: true}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 30 * Second, Nil: true}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 40 * Second, Nil: true}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 50 * Second, Nil: true}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=B"), Time: 0 * Second, Nil: true}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=B"), Time: 10 * Second, Nil: true}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=B"), Time: 20 * Second, Nil: true}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=B"), Time: 30 * Second, Value: 4, Aggregated: 1}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=B"), Time: 40 * Second, Nil: true}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=B"), Time: 50 * Second, Nil: true}},
			},
		},
		{
			name: "Fill_Linear_Integer_One",
			q:    `SELECT max(value) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-01T00:01:00Z' GROUP BY host, time(10s) fill(linear)`,
			typ:  influxql.Integer,
			expr: `max(value::integer)`,
			itrs: []query.Iterator{
				&IntegerIterator{Points: []query.IntegerPoint{
					{Name: "cpu", Tags: ParseTags("host=A"), Time: 12 * Second, Value: 1},
					{Name: "cpu", Tags: ParseTags("host=A"), Time: 32 * Second, Value: 4},
				}},
			},
			points: [][]query.Point{
				{&query.IntegerPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 0 * Second, Nil: true}},
				{&query.IntegerPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 10 * Second, Value: 1, Aggregated: 1}},
				{&query.IntegerPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 20 * Second, Value: 2}},
				{&query.IntegerPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 30 * Second, Value: 4, Aggregated: 1}},
				{&query.IntegerPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 40 * Second, Nil: true}},
				{&query.IntegerPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 50 * Second, Nil: true}},
			},
		},
		{
			name: "Fill_Linear_Integer_Many",
			q:    `SELECT max(value) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-01T00:01:20Z' GROUP BY host, time(10s) fill(linear)`,
			typ:  influxql.Integer,
			expr: `max(value::integer)`,
			itrs: []query.Iterator{
				&IntegerIterator{Points: []query.IntegerPoint{
					{Name: "cpu", Tags: ParseTags("host=A"), Time: 12 * Second, Value: 1},
					{Name: "cpu", Tags: ParseTags("host=A"), Time: 72 * Second, Value: 10},
				}},
			},
			points: [][]query.Point{
				{&query.IntegerPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 0 * Second, Nil: true}},
				{&query.IntegerPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 10 * Second, Value: 1, Aggregated: 1}},
				{&query.IntegerPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 20 * Second, Value: 2}},
				{&query.IntegerPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 30 * Second, Value: 4}},
				{&query.IntegerPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 40 * Second, Value: 5}},
				{&query.IntegerPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 50 * Second, Value: 7}},
				{&query.IntegerPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 60 * Second, Value: 8}},
				{&query.IntegerPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 70 * Second, Value: 10, Aggregated: 1}},
			},
		},
		{
			name: "Fill_Linear_Integer_MultipleSeries",
			q:    `SELECT max(value) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-01T00:01:00Z' GROUP BY host, time(10s) fill(linear)`,
			typ:  influxql.Integer,
			expr: `max(value::integer)`,
			itrs: []query.Iterator{
				&IntegerIterator{Points: []query.IntegerPoint{
					{Name: "cpu", Tags: ParseTags("host=A"), Time: 12 * Second, Value: 2},
					{Name: "cpu", Tags: ParseTags("host=B"), Time: 32 * Second, Value: 4},
				}},
			},
			points: [][]query.Point{
				{&query.IntegerPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 0 * Second, Nil: true}},
				{&query.IntegerPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 10 * Second, Value: 2, Aggregated: 1}},
				{&query.IntegerPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 20 * Second, Nil: true}},
				{&query.IntegerPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 30 * Second, Nil: true}},
				{&query.IntegerPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 40 * Second, Nil: true}},
				{&query.IntegerPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 50 * Second, Nil: true}},
				{&query.IntegerPoint{Name: "cpu", Tags: ParseTags("host=B"), Time: 0 * Second, Nil: true}},
				{&query.IntegerPoint{Name: "cpu", Tags: ParseTags("host=B"), Time: 10 * Second, Nil: true}},
				{&query.IntegerPoint{Name: "cpu", Tags: ParseTags("host=B"), Time: 20 * Second, Nil: true}},
				{&query.IntegerPoint{Name: "cpu", Tags: ParseTags("host=B"), Time: 30 * Second, Value: 4, Aggregated: 1}},
				{&query.IntegerPoint{Name: "cpu", Tags: ParseTags("host=B"), Time: 40 * Second, Nil: true}},
				{&query.IntegerPoint{Name: "cpu", Tags: ParseTags("host=B"), Time: 50 * Second, Nil: true}},
			},
		},
		{
			name: "Stddev_Float",
			q:    `SELECT stddev(value) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-02T00:00:00Z' GROUP BY time(10s), host fill(none)`,
			typ:  influxql.Float,
			itrs: []query.Iterator{
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 0 * Second, Value: 20},
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 11 * Second, Value: 3},
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 31 * Second, Value: 100},
				}},
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 5 * Second, Value: 10},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 50 * Second, Value: 1},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 51 * Second, Value: 2},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 52 * Second, Value: 3},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 53 * Second, Value: 4},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 53 * Second, Value: 5},
				}},
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 9 * Second, Value: 19},
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 10 * Second, Value: 2},
				}},
			},
			points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 0 * Second, Value: 0.7071067811865476}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 10 * Second, Value: 0.7071067811865476}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 30 * Second, Nil: true}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=B"), Time: 0 * Second, Nil: true}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=B"), Time: 50 * Second, Value: 1.5811388300841898}},
			},
		},
		{
			name: "Stddev_Integer",
			q:    `SELECT stddev(value) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-02T00:00:00Z' GROUP BY time(10s), host fill(none)`,
			typ:  influxql.Integer,
			itrs: []query.Iterator{
				&IntegerIterator{Points: []query.IntegerPoint{
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 0 * Second, Value: 20},
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 11 * Second, Value: 3},
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 31 * Second, Value: 100},
				}},
				&IntegerIterator{Points: []query.IntegerPoint{
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 5 * Second, Value: 10},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 50 * Second, Value: 1},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 51 * Second, Value: 2},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 52 * Second, Value: 3},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 53 * Second, Value: 4},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 53 * Second, Value: 5},
				}},
				&IntegerIterator{Points: []query.IntegerPoint{
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 9 * Second, Value: 19},
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 10 * Second, Value: 2},
				}},
			},
			points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 0 * Second, Value: 0.7071067811865476}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 10 * Second, Value: 0.7071067811865476}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 30 * Second, Nil: true}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=B"), Time: 0 * Second, Nil: true}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=B"), Time: 50 * Second, Value: 1.5811388300841898}},
			},
		},
		{
			name: "Spread_Float",
			q:    `SELECT spread(value) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-02T00:00:00Z' GROUP BY time(10s), host fill(none)`,
			typ:  influxql.Float,
			itrs: []query.Iterator{
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 0 * Second, Value: 20},
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 11 * Second, Value: 3},
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 31 * Second, Value: 100},
				}},
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 5 * Second, Value: 10},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 50 * Second, Value: 1},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 51 * Second, Value: 2},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 52 * Second, Value: 3},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 53 * Second, Value: 4},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 53 * Second, Value: 5},
				}},
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 9 * Second, Value: 19},
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 10 * Second, Value: 2},
				}},
			},
			points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 0 * Second, Value: 1}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 10 * Second, Value: 1}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 30 * Second, Value: 0}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=B"), Time: 0 * Second, Value: 0}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=B"), Time: 50 * Second, Value: 4}},
			},
		},
		{
			name: "Spread_Integer",
			q:    `SELECT spread(value) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-02T00:00:00Z' GROUP BY time(10s), host fill(none)`,
			typ:  influxql.Integer,
			itrs: []query.Iterator{
				&IntegerIterator{Points: []query.IntegerPoint{
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 0 * Second, Value: 20},
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 11 * Second, Value: 3},
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 31 * Second, Value: 100},
				}},
				&IntegerIterator{Points: []query.IntegerPoint{
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 5 * Second, Value: 10},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 50 * Second, Value: 1},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 51 * Second, Value: 2},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 52 * Second, Value: 3},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 53 * Second, Value: 4},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 53 * Second, Value: 5},
				}},
				&IntegerIterator{Points: []query.IntegerPoint{
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 9 * Second, Value: 19},
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 10 * Second, Value: 2},
				}},
			},
			points: [][]query.Point{
				{&query.IntegerPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 0 * Second, Value: 1}},
				{&query.IntegerPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 10 * Second, Value: 1}},
				{&query.IntegerPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 30 * Second, Value: 0}},
				{&query.IntegerPoint{Name: "cpu", Tags: ParseTags("host=B"), Time: 0 * Second, Value: 0}},
				{&query.IntegerPoint{Name: "cpu", Tags: ParseTags("host=B"), Time: 50 * Second, Value: 4}},
			},
		},
		{
			name: "Percentile_Float",
			q:    `SELECT percentile(value, 90) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-02T00:00:00Z' GROUP BY time(10s), host fill(none)`,
			typ:  influxql.Float,
			itrs: []query.Iterator{
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 0 * Second, Value: 20},
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 11 * Second, Value: 3},
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 31 * Second, Value: 100},
				}},
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 5 * Second, Value: 10},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 50 * Second, Value: 10},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 51 * Second, Value: 9},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 52 * Second, Value: 8},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 53 * Second, Value: 7},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 54 * Second, Value: 6},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 55 * Second, Value: 5},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 56 * Second, Value: 4},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 57 * Second, Value: 3},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 58 * Second, Value: 2},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 59 * Second, Value: 1},
				}},
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 9 * Second, Value: 19},
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 10 * Second, Value: 2},
				}},
			},
			points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 0 * Second, Value: 20}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 10 * Second, Value: 3}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 30 * Second, Value: 100}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=B"), Time: 0 * Second, Value: 10}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=B"), Time: 50 * Second, Value: 9}},
			},
		},
		{
			name: "Percentile_Integer",
			q:    `SELECT percentile(value, 90) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-02T00:00:00Z' GROUP BY time(10s), host fill(none)`,
			typ:  influxql.Integer,
			itrs: []query.Iterator{
				&IntegerIterator{Points: []query.IntegerPoint{
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 0 * Second, Value: 20},
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 11 * Second, Value: 3},
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 31 * Second, Value: 100},
				}},
				&IntegerIterator{Points: []query.IntegerPoint{
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 5 * Second, Value: 10},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 50 * Second, Value: 10},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 51 * Second, Value: 9},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 52 * Second, Value: 8},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 53 * Second, Value: 7},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 54 * Second, Value: 6},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 55 * Second, Value: 5},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 56 * Second, Value: 4},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 57 * Second, Value: 3},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 58 * Second, Value: 2},
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 59 * Second, Value: 1},
				}},
				&IntegerIterator{Points: []query.IntegerPoint{
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 9 * Second, Value: 19},
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 10 * Second, Value: 2},
				}},
			},
			points: [][]query.Point{
				{&query.IntegerPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 0 * Second, Value: 20}},
				{&query.IntegerPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 10 * Second, Value: 3}},
				{&query.IntegerPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 30 * Second, Value: 100}},
				{&query.IntegerPoint{Name: "cpu", Tags: ParseTags("host=B"), Time: 0 * Second, Value: 10}},
				{&query.IntegerPoint{Name: "cpu", Tags: ParseTags("host=B"), Time: 50 * Second, Value: 9}},
			},
		},
		{
			name: "Sample_Float",
			q:    `SELECT sample(value, 2) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-02T00:00:00Z' GROUP BY time(10s), host fill(none)`,
			typ:  influxql.Float,
			itrs: []query.Iterator{
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 0 * Second, Value: 20},
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 5 * Second, Value: 10},
				}},
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Tags: ParseTags("region=east,host=B"), Time: 10 * Second, Value: 19},
					{Name: "cpu", Tags: ParseTags("region=east,host=B"), Time: 15 * Second, Value: 2},
				}},
			},
			points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 0 * Second, Value: 20}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 5 * Second, Value: 10}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=B"), Time: 10 * Second, Value: 19}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=B"), Time: 15 * Second, Value: 2}},
			},
		},
		{
			name: "Sample_Integer",
			q:    `SELECT sample(value, 2) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-02T00:00:00Z' GROUP BY time(10s), host fill(none)`,
			typ:  influxql.Integer,
			itrs: []query.Iterator{
				&IntegerIterator{Points: []query.IntegerPoint{
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 0 * Second, Value: 20},
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 5 * Second, Value: 10},
				}},
				&IntegerIterator{Points: []query.IntegerPoint{
					{Name: "cpu", Tags: ParseTags("region=east,host=B"), Time: 10 * Second, Value: 19},
					{Name: "cpu", Tags: ParseTags("region=east,host=B"), Time: 15 * Second, Value: 2},
				}},
			},
			points: [][]query.Point{
				{&query.IntegerPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 0 * Second, Value: 20}},
				{&query.IntegerPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 5 * Second, Value: 10}},
				{&query.IntegerPoint{Name: "cpu", Tags: ParseTags("host=B"), Time: 10 * Second, Value: 19}},
				{&query.IntegerPoint{Name: "cpu", Tags: ParseTags("host=B"), Time: 15 * Second, Value: 2}},
			},
		},
		{
			name: "Sample_String",
			q:    `SELECT sample(value, 2) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-02T00:00:00Z' GROUP BY time(10s), host fill(none)`,
			typ:  influxql.String,
			itrs: []query.Iterator{
				&StringIterator{Points: []query.StringPoint{
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 0 * Second, Value: "a"},
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 5 * Second, Value: "b"},
				}},
				&StringIterator{Points: []query.StringPoint{
					{Name: "cpu", Tags: ParseTags("region=east,host=B"), Time: 10 * Second, Value: "c"},
					{Name: "cpu", Tags: ParseTags("region=east,host=B"), Time: 15 * Second, Value: "d"},
				}},
			},
			points: [][]query.Point{
				{&query.StringPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 0 * Second, Value: "a"}},
				{&query.StringPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 5 * Second, Value: "b"}},
				{&query.StringPoint{Name: "cpu", Tags: ParseTags("host=B"), Time: 10 * Second, Value: "c"}},
				{&query.StringPoint{Name: "cpu", Tags: ParseTags("host=B"), Time: 15 * Second, Value: "d"}},
			},
		},
		{
			name: "Sample_Boolean",
			q:    `SELECT sample(value, 2) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-02T00:00:00Z' GROUP BY time(10s), host fill(none)`,
			typ:  influxql.Boolean,
			itrs: []query.Iterator{
				&BooleanIterator{Points: []query.BooleanPoint{
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 0 * Second, Value: true},
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 5 * Second, Value: false},
				}},
				&BooleanIterator{Points: []query.BooleanPoint{
					{Name: "cpu", Tags: ParseTags("region=east,host=B"), Time: 10 * Second, Value: false},
					{Name: "cpu", Tags: ParseTags("region=east,host=B"), Time: 15 * Second, Value: true},
				}},
			},
			points: [][]query.Point{
				{&query.BooleanPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 0 * Second, Value: true}},
				{&query.BooleanPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 5 * Second, Value: false}},
				{&query.BooleanPoint{Name: "cpu", Tags: ParseTags("host=B"), Time: 10 * Second, Value: false}},
				{&query.BooleanPoint{Name: "cpu", Tags: ParseTags("host=B"), Time: 15 * Second, Value: true}},
			},
		},
		//{
		//	name: "Raw",
		//	q:    `SELECT v1::float, v2::float FROM cpu`,
		//	itrs: []query.Iterator{
		//		&FloatIterator{Points: []query.FloatPoint{
		//			{Time: 0, Aux: []interface{}{float64(1), nil}},
		//			{Time: 1, Aux: []interface{}{nil, float64(2)}},
		//			{Time: 5, Aux: []interface{}{float64(3), float64(4)}},
		//		}},
		//	},
		//	points: [][]query.Point{
		//		{
		//			&query.FloatPoint{Time: 0, Value: 1},
		//			&query.FloatPoint{Time: 0, Nil: true},
		//		},
		//		{
		//			&query.FloatPoint{Time: 1, Nil: true},
		//			&query.FloatPoint{Time: 1, Value: 2},
		//		},
		//		{
		//			&query.FloatPoint{Time: 5, Value: 3},
		//			&query.FloatPoint{Time: 5, Value: 4},
		//		},
		//	},
		//},
		{
			name: "ParenExpr_Min",
			q:    `SELECT (min(value)) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-02T00:00:00Z' GROUP BY time(10s), host fill(none)`,
			typ:  influxql.Float,
			expr: `min(value::float)`,
			itrs: []query.Iterator{
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 0 * Second, Value: 20},
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 11 * Second, Value: 3},
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 31 * Second, Value: 100},
				}},
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 5 * Second, Value: 10},
				}},
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 9 * Second, Value: 19},
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 10 * Second, Value: 2},
				}},
			},
			points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 0 * Second, Value: 19, Aggregated: 2}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 10 * Second, Value: 2, Aggregated: 2}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 30 * Second, Value: 100, Aggregated: 1}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=B"), Time: 0 * Second, Value: 10, Aggregated: 1}},
			},
		},
		{
			name: "ParenExpr_Distinct",
			q:    `SELECT (distinct(value)) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-02T00:00:00Z' GROUP BY time(10s), host fill(none)`,
			typ:  influxql.Float,
			itrs: []query.Iterator{
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 0 * Second, Value: 20},
					{Name: "cpu", Tags: ParseTags("region=west,host=A"), Time: 1 * Second, Value: 19},
				}},
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Tags: ParseTags("region=west,host=B"), Time: 5 * Second, Value: 10},
				}},
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 9 * Second, Value: 19},
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 10 * Second, Value: 2},
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 11 * Second, Value: 2},
					{Name: "cpu", Tags: ParseTags("region=east,host=A"), Time: 12 * Second, Value: 2},
				}},
			},
			points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 0 * Second, Value: 20}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 0 * Second, Value: 19}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=A"), Time: 10 * Second, Value: 2}},
				{&query.FloatPoint{Name: "cpu", Tags: ParseTags("host=B"), Time: 0 * Second, Value: 10}},
			},
		},
		{
			name: "Derivative_Float",
			q:    `SELECT derivative(value, 1s) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-01T00:00:16Z'`,
			typ:  influxql.Float,
			itrs: []query.Iterator{
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Time: 0 * Second, Value: 20},
					{Name: "cpu", Time: 4 * Second, Value: 10},
					{Name: "cpu", Time: 8 * Second, Value: 19},
					{Name: "cpu", Time: 12 * Second, Value: 3},
				}},
			},
			points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Time: 4 * Second, Value: -2.5}},
				{&query.FloatPoint{Name: "cpu", Time: 8 * Second, Value: 2.25}},
				{&query.FloatPoint{Name: "cpu", Time: 12 * Second, Value: -4}},
			},
		},
		{
			name: "Derivative_Integer",
			q:    `SELECT derivative(value, 1s) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-01T00:00:16Z'`,
			typ:  influxql.Integer,
			itrs: []query.Iterator{
				&IntegerIterator{Points: []query.IntegerPoint{
					{Name: "cpu", Time: 0 * Second, Value: 20},
					{Name: "cpu", Time: 4 * Second, Value: 10},
					{Name: "cpu", Time: 8 * Second, Value: 19},
					{Name: "cpu", Time: 12 * Second, Value: 3},
				}},
			},
			points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Time: 4 * Second, Value: -2.5}},
				{&query.FloatPoint{Name: "cpu", Time: 8 * Second, Value: 2.25}},
				{&query.FloatPoint{Name: "cpu", Time: 12 * Second, Value: -4}},
			},
		},
		{
			name: "Derivative_Desc_Float",
			q:    `SELECT derivative(value, 1s) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-01T00:00:16Z' ORDER BY desc`,
			typ:  influxql.Float,
			itrs: []query.Iterator{
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Time: 12 * Second, Value: 3},
					{Name: "cpu", Time: 8 * Second, Value: 19},
					{Name: "cpu", Time: 4 * Second, Value: 10},
					{Name: "cpu", Time: 0 * Second, Value: 20},
				}},
			},
			points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Time: 8 * Second, Value: 4}},
				{&query.FloatPoint{Name: "cpu", Time: 4 * Second, Value: -2.25}},
				{&query.FloatPoint{Name: "cpu", Time: 0 * Second, Value: 2.5}},
			},
		},
		{
			name: "Derivative_Desc_Integer",
			q:    `SELECT derivative(value, 1s) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-01T00:00:16Z' ORDER BY desc`,
			typ:  influxql.Integer,
			itrs: []query.Iterator{
				&IntegerIterator{Points: []query.IntegerPoint{
					{Name: "cpu", Time: 12 * Second, Value: 3},
					{Name: "cpu", Time: 8 * Second, Value: 19},
					{Name: "cpu", Time: 4 * Second, Value: 10},
					{Name: "cpu", Time: 0 * Second, Value: 20},
				}},
			},
			points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Time: 8 * Second, Value: 4}},
				{&query.FloatPoint{Name: "cpu", Time: 4 * Second, Value: -2.25}},
				{&query.FloatPoint{Name: "cpu", Time: 0 * Second, Value: 2.5}},
			},
		},
		{
			name: "Derivative_Duplicate_Float",
			q:    `SELECT derivative(value, 1s) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-01T00:00:16Z'`,
			typ:  influxql.Float,
			itrs: []query.Iterator{
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Time: 0 * Second, Value: 20},
					{Name: "cpu", Time: 0 * Second, Value: 19},
					{Name: "cpu", Time: 4 * Second, Value: 10},
					{Name: "cpu", Time: 4 * Second, Value: 3},
				}},
			},
			points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Time: 4 * Second, Value: -2.5}},
			},
		},
		{
			name: "Derivative_Duplicate_Integer",
			q:    `SELECT derivative(value, 1s) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-01T00:00:16Z'`,
			typ:  influxql.Integer,
			itrs: []query.Iterator{
				&IntegerIterator{Points: []query.IntegerPoint{
					{Name: "cpu", Time: 0 * Second, Value: 20},
					{Name: "cpu", Time: 0 * Second, Value: 19},
					{Name: "cpu", Time: 4 * Second, Value: 10},
					{Name: "cpu", Time: 4 * Second, Value: 3},
				}},
			},
			points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Time: 4 * Second, Value: -2.5}},
			},
		},
		{
			name: "Difference_Float",
			q:    `SELECT difference(value) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-01T00:00:16Z'`,
			typ:  influxql.Float,
			itrs: []query.Iterator{
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Time: 0 * Second, Value: 20},
					{Name: "cpu", Time: 4 * Second, Value: 10},
					{Name: "cpu", Time: 8 * Second, Value: 19},
					{Name: "cpu", Time: 12 * Second, Value: 3},
				}},
			},
			points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Time: 4 * Second, Value: -10}},
				{&query.FloatPoint{Name: "cpu", Time: 8 * Second, Value: 9}},
				{&query.FloatPoint{Name: "cpu", Time: 12 * Second, Value: -16}},
			},
		},
		{
			name: "Difference_Integer",
			q:    `SELECT difference(value) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-01T00:00:16Z'`,
			typ:  influxql.Integer,
			itrs: []query.Iterator{
				&IntegerIterator{Points: []query.IntegerPoint{
					{Name: "cpu", Time: 0 * Second, Value: 20},
					{Name: "cpu", Time: 4 * Second, Value: 10},
					{Name: "cpu", Time: 8 * Second, Value: 19},
					{Name: "cpu", Time: 12 * Second, Value: 3},
				}},
			},
			points: [][]query.Point{
				{&query.IntegerPoint{Name: "cpu", Time: 4 * Second, Value: -10}},
				{&query.IntegerPoint{Name: "cpu", Time: 8 * Second, Value: 9}},
				{&query.IntegerPoint{Name: "cpu", Time: 12 * Second, Value: -16}},
			},
		},
		{
			name: "Difference_Duplicate_Float",
			q:    `SELECT difference(value) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-01T00:00:16Z'`,
			typ:  influxql.Float,
			itrs: []query.Iterator{
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Time: 0 * Second, Value: 20},
					{Name: "cpu", Time: 0 * Second, Value: 19},
					{Name: "cpu", Time: 4 * Second, Value: 10},
					{Name: "cpu", Time: 4 * Second, Value: 3},
				}},
			},
			points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Time: 4 * Second, Value: -10}},
			},
		},
		{
			name: "Difference_Duplicate_Integer",
			q:    `SELECT difference(value) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-01T00:00:16Z'`,
			typ:  influxql.Integer,
			itrs: []query.Iterator{
				&IntegerIterator{Points: []query.IntegerPoint{
					{Name: "cpu", Time: 0 * Second, Value: 20},
					{Name: "cpu", Time: 0 * Second, Value: 19},
					{Name: "cpu", Time: 4 * Second, Value: 10},
					{Name: "cpu", Time: 4 * Second, Value: 3},
				}},
			},
			points: [][]query.Point{
				{&query.IntegerPoint{Name: "cpu", Time: 4 * Second, Value: -10}},
			},
		},
		{
			name: "Non_Negative_Difference_Float",
			q:    `SELECT non_negative_difference(value) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-01T00:00:16Z'`,
			typ:  influxql.Float,
			itrs: []query.Iterator{
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Time: 0 * Second, Value: 20},
					{Name: "cpu", Time: 4 * Second, Value: 10},
					{Name: "cpu", Time: 8 * Second, Value: 29},
					{Name: "cpu", Time: 12 * Second, Value: 3},
					{Name: "cpu", Time: 16 * Second, Value: 39},
				}},
			},
			points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Time: 8 * Second, Value: 19}},
				{&query.FloatPoint{Name: "cpu", Time: 16 * Second, Value: 36}},
			},
		},
		{
			name: "Non_Negative_Difference_Integer",
			q:    `SELECT non_negative_difference(value) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-01T00:00:16Z'`,
			typ:  influxql.Integer,
			itrs: []query.Iterator{
				&IntegerIterator{Points: []query.IntegerPoint{
					{Name: "cpu", Time: 0 * Second, Value: 20},
					{Name: "cpu", Time: 4 * Second, Value: 10},
					{Name: "cpu", Time: 8 * Second, Value: 21},
					{Name: "cpu", Time: 12 * Second, Value: 3},
				}},
			},
			points: [][]query.Point{
				{&query.IntegerPoint{Name: "cpu", Time: 8 * Second, Value: 11}},
			},
		},
		{
			name: "Non_Negative_Difference_Duplicate_Float",
			q:    `SELECT non_negative_difference(value) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-01T00:00:16Z'`,
			typ:  influxql.Float,
			itrs: []query.Iterator{
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Time: 0 * Second, Value: 20},
					{Name: "cpu", Time: 0 * Second, Value: 19},
					{Name: "cpu", Time: 4 * Second, Value: 10},
					{Name: "cpu", Time: 4 * Second, Value: 3},
					{Name: "cpu", Time: 8 * Second, Value: 30},
					{Name: "cpu", Time: 8 * Second, Value: 19},
					{Name: "cpu", Time: 12 * Second, Value: 10},
					{Name: "cpu", Time: 12 * Second, Value: 3},
					{Name: "cpu", Time: 16 * Second, Value: 40},
					{Name: "cpu", Time: 16 * Second, Value: 3},
				}},
			},
			points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Time: 8 * Second, Value: 20}},
				{&query.FloatPoint{Name: "cpu", Time: 16 * Second, Value: 30}},
			},
		},
		{
			name: "Non_Negative_Difference_Duplicate_Integer",
			q:    `SELECT non_negative_difference(value) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-01T00:00:16Z'`,
			typ:  influxql.Integer,
			itrs: []query.Iterator{
				&IntegerIterator{Points: []query.IntegerPoint{
					{Name: "cpu", Time: 0 * Second, Value: 20},
					{Name: "cpu", Time: 0 * Second, Value: 19},
					{Name: "cpu", Time: 4 * Second, Value: 10},
					{Name: "cpu", Time: 4 * Second, Value: 3},
					{Name: "cpu", Time: 8 * Second, Value: 30},
					{Name: "cpu", Time: 8 * Second, Value: 19},
					{Name: "cpu", Time: 12 * Second, Value: 10},
					{Name: "cpu", Time: 12 * Second, Value: 3},
					{Name: "cpu", Time: 16 * Second, Value: 40},
					{Name: "cpu", Time: 16 * Second, Value: 3},
				}},
			},
			points: [][]query.Point{
				{&query.IntegerPoint{Name: "cpu", Time: 8 * Second, Value: 20}},
				{&query.IntegerPoint{Name: "cpu", Time: 16 * Second, Value: 30}},
			},
		},
		{
			name: "Elapsed_Float",
			q:    `SELECT elapsed(value, 1s) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-01T00:00:16Z'`,
			typ:  influxql.Float,
			itrs: []query.Iterator{
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Time: 0 * Second, Value: 20},
					{Name: "cpu", Time: 4 * Second, Value: 10},
					{Name: "cpu", Time: 8 * Second, Value: 19},
					{Name: "cpu", Time: 11 * Second, Value: 3},
				}},
			},
			points: [][]query.Point{
				{&query.IntegerPoint{Name: "cpu", Time: 4 * Second, Value: 4}},
				{&query.IntegerPoint{Name: "cpu", Time: 8 * Second, Value: 4}},
				{&query.IntegerPoint{Name: "cpu", Time: 11 * Second, Value: 3}},
			},
		},
		{
			name: "Elapsed_Integer",
			q:    `SELECT elapsed(value, 1s) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-01T00:00:16Z'`,
			typ:  influxql.Integer,
			itrs: []query.Iterator{
				&IntegerIterator{Points: []query.IntegerPoint{
					{Name: "cpu", Time: 0 * Second, Value: 20},
					{Name: "cpu", Time: 4 * Second, Value: 10},
					{Name: "cpu", Time: 8 * Second, Value: 19},
					{Name: "cpu", Time: 11 * Second, Value: 3},
				}},
			},
			points: [][]query.Point{
				{&query.IntegerPoint{Name: "cpu", Time: 4 * Second, Value: 4}},
				{&query.IntegerPoint{Name: "cpu", Time: 8 * Second, Value: 4}},
				{&query.IntegerPoint{Name: "cpu", Time: 11 * Second, Value: 3}},
			},
		},
		{
			name: "Elapsed_String",
			q:    `SELECT elapsed(value, 1s) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-01T00:00:16Z'`,
			typ:  influxql.String,
			itrs: []query.Iterator{
				&StringIterator{Points: []query.StringPoint{
					{Name: "cpu", Time: 0 * Second, Value: "a"},
					{Name: "cpu", Time: 4 * Second, Value: "b"},
					{Name: "cpu", Time: 8 * Second, Value: "c"},
					{Name: "cpu", Time: 11 * Second, Value: "d"},
				}},
			},
			points: [][]query.Point{
				{&query.IntegerPoint{Name: "cpu", Time: 4 * Second, Value: 4}},
				{&query.IntegerPoint{Name: "cpu", Time: 8 * Second, Value: 4}},
				{&query.IntegerPoint{Name: "cpu", Time: 11 * Second, Value: 3}},
			},
		},
		{
			name: "Elapsed_Boolean",
			q:    `SELECT elapsed(value, 1s) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-01T00:00:16Z'`,
			typ:  influxql.Boolean,
			itrs: []query.Iterator{
				&BooleanIterator{Points: []query.BooleanPoint{
					{Name: "cpu", Time: 0 * Second, Value: true},
					{Name: "cpu", Time: 4 * Second, Value: false},
					{Name: "cpu", Time: 8 * Second, Value: false},
					{Name: "cpu", Time: 11 * Second, Value: true},
				}},
			},
			points: [][]query.Point{
				{&query.IntegerPoint{Name: "cpu", Time: 4 * Second, Value: 4}},
				{&query.IntegerPoint{Name: "cpu", Time: 8 * Second, Value: 4}},
				{&query.IntegerPoint{Name: "cpu", Time: 11 * Second, Value: 3}},
			},
		},
		{
			name: "Integral_Float",
			q:    `SELECT integral(value) FROM cpu`,
			typ:  influxql.Float,
			itrs: []query.Iterator{
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Time: 10 * Second, Value: 20},
					{Name: "cpu", Time: 15 * Second, Value: 10},
					{Name: "cpu", Time: 20 * Second, Value: 0},
					{Name: "cpu", Time: 30 * Second, Value: -10},
				}},
			},
			points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Time: 0, Value: 50}},
			},
		},
		{
			name: "Integral_Duplicate_Float",
			q:    `SELECT integral(value) FROM cpu`,
			typ:  influxql.Float,
			itrs: []query.Iterator{
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Time: 0 * Second, Value: 20},
					{Name: "cpu", Time: 5 * Second, Value: 10},
					{Name: "cpu", Time: 5 * Second, Value: 30},
					{Name: "cpu", Time: 10 * Second, Value: 40},
				}},
			},
			points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Time: 0, Value: 250}},
			},
		},
		{
			name: "Integral_Float_GroupByTime",
			q:    `SELECT integral(value) FROM cpu WHERE time > 0s AND time < 60s GROUP BY time(20s)`,
			typ:  influxql.Float,
			itrs: []query.Iterator{
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Time: 10 * Second, Value: 20},
					{Name: "cpu", Time: 15 * Second, Value: 10},
					{Name: "cpu", Time: 20 * Second, Value: 0},
					{Name: "cpu", Time: 30 * Second, Value: -10},
				}},
			},
			points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Time: 0, Value: 100}},
				{&query.FloatPoint{Name: "cpu", Time: 20 * Second, Value: -50}},
			},
		},
		{
			name: "Integral_Float_InterpolateGroupByTime",
			q:    `SELECT integral(value) FROM cpu WHERE time > 0s AND time < 60s GROUP BY time(20s)`,
			typ:  influxql.Float,
			itrs: []query.Iterator{
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Time: 10 * Second, Value: 20},
					{Name: "cpu", Time: 15 * Second, Value: 10},
					{Name: "cpu", Time: 25 * Second, Value: 0},
					{Name: "cpu", Time: 30 * Second, Value: -10},
				}},
			},
			points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Time: 0, Value: 112.5}},
				{&query.FloatPoint{Name: "cpu", Time: 20 * Second, Value: -12.5}},
			},
		},
		{
			name: "Integral_Integer",
			q:    `SELECT integral(value) FROM cpu`,
			typ:  influxql.Integer,
			itrs: []query.Iterator{
				&IntegerIterator{Points: []query.IntegerPoint{
					{Name: "cpu", Time: 0 * Second, Value: 20},
					{Name: "cpu", Time: 5 * Second, Value: 10},
					{Name: "cpu", Time: 10 * Second, Value: 0},
					{Name: "cpu", Time: 20 * Second, Value: -10},
				}},
			},
			points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Time: 0, Value: 50}},
			},
		},
		{
			name: "Integral_Duplicate_Integer",
			q:    `SELECT integral(value, 2s) FROM cpu`,
			typ:  influxql.Integer,
			itrs: []query.Iterator{
				&IntegerIterator{Points: []query.IntegerPoint{
					{Name: "cpu", Time: 0 * Second, Value: 20},
					{Name: "cpu", Time: 5 * Second, Value: 10},
					{Name: "cpu", Time: 5 * Second, Value: 30},
					{Name: "cpu", Time: 10 * Second, Value: 40},
				}},
			},
			points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Time: 0, Value: 125}},
			},
		},
		{
			name: "MovingAverage_Float",
			q:    `SELECT moving_average(value, 2) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-01T00:00:16Z'`,
			typ:  influxql.Float,
			itrs: []query.Iterator{
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Time: 0 * Second, Value: 20},
					{Name: "cpu", Time: 4 * Second, Value: 10},
					{Name: "cpu", Time: 8 * Second, Value: 19},
					{Name: "cpu", Time: 12 * Second, Value: 3},
				}},
			},
			points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Time: 4 * Second, Value: 15, Aggregated: 2}},
				{&query.FloatPoint{Name: "cpu", Time: 8 * Second, Value: 14.5, Aggregated: 2}},
				{&query.FloatPoint{Name: "cpu", Time: 12 * Second, Value: 11, Aggregated: 2}},
			},
		},
		{
			name: "MovingAverage_Integer",
			q:    `SELECT moving_average(value, 2) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-01T00:00:16Z'`,
			typ:  influxql.Integer,
			itrs: []query.Iterator{
				&IntegerIterator{Points: []query.IntegerPoint{
					{Name: "cpu", Time: 0 * Second, Value: 20},
					{Name: "cpu", Time: 4 * Second, Value: 10},
					{Name: "cpu", Time: 8 * Second, Value: 19},
					{Name: "cpu", Time: 12 * Second, Value: 3},
				}},
			},
			points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Time: 4 * Second, Value: 15, Aggregated: 2}},
				{&query.FloatPoint{Name: "cpu", Time: 8 * Second, Value: 14.5, Aggregated: 2}},
				{&query.FloatPoint{Name: "cpu", Time: 12 * Second, Value: 11, Aggregated: 2}},
			},
		},
		{
			name: "CumulativeSum_Float",
			q:    `SELECT cumulative_sum(value) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-01T00:00:16Z'`,
			typ:  influxql.Float,
			itrs: []query.Iterator{
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Time: 0 * Second, Value: 20},
					{Name: "cpu", Time: 4 * Second, Value: 10},
					{Name: "cpu", Time: 8 * Second, Value: 19},
					{Name: "cpu", Time: 12 * Second, Value: 3},
				}},
			},
			points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Time: 0 * Second, Value: 20}},
				{&query.FloatPoint{Name: "cpu", Time: 4 * Second, Value: 30}},
				{&query.FloatPoint{Name: "cpu", Time: 8 * Second, Value: 49}},
				{&query.FloatPoint{Name: "cpu", Time: 12 * Second, Value: 52}},
			},
		},
		{
			name: "CumulativeSum_Integer",
			q:    `SELECT cumulative_sum(value) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-01T00:00:16Z'`,
			typ:  influxql.Integer,
			itrs: []query.Iterator{
				&IntegerIterator{Points: []query.IntegerPoint{
					{Name: "cpu", Time: 0 * Second, Value: 20},
					{Name: "cpu", Time: 4 * Second, Value: 10},
					{Name: "cpu", Time: 8 * Second, Value: 19},
					{Name: "cpu", Time: 12 * Second, Value: 3},
				}},
			},
			points: [][]query.Point{
				{&query.IntegerPoint{Name: "cpu", Time: 0 * Second, Value: 20}},
				{&query.IntegerPoint{Name: "cpu", Time: 4 * Second, Value: 30}},
				{&query.IntegerPoint{Name: "cpu", Time: 8 * Second, Value: 49}},
				{&query.IntegerPoint{Name: "cpu", Time: 12 * Second, Value: 52}},
			},
		},
		{
			name: "CumulativeSum_Duplicate_Float",
			q:    `SELECT cumulative_sum(value) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-01T00:00:16Z'`,
			typ:  influxql.Float,
			itrs: []query.Iterator{
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Time: 0 * Second, Value: 20},
					{Name: "cpu", Time: 0 * Second, Value: 19},
					{Name: "cpu", Time: 4 * Second, Value: 10},
					{Name: "cpu", Time: 4 * Second, Value: 3},
				}},
			},
			points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Time: 0 * Second, Value: 20}},
				{&query.FloatPoint{Name: "cpu", Time: 0 * Second, Value: 39}},
				{&query.FloatPoint{Name: "cpu", Time: 4 * Second, Value: 49}},
				{&query.FloatPoint{Name: "cpu", Time: 4 * Second, Value: 52}},
			},
		},
		{
			name: "CumulativeSum_Duplicate_Integer",
			q:    `SELECT cumulative_sum(value) FROM cpu WHERE time >= '1970-01-01T00:00:00Z' AND time < '1970-01-01T00:00:16Z'`,
			typ:  influxql.Integer,
			itrs: []query.Iterator{
				&IntegerIterator{Points: []query.IntegerPoint{
					{Name: "cpu", Time: 0 * Second, Value: 20},
					{Name: "cpu", Time: 0 * Second, Value: 19},
					{Name: "cpu", Time: 4 * Second, Value: 10},
					{Name: "cpu", Time: 4 * Second, Value: 3},
				}},
			},
			points: [][]query.Point{
				{&query.IntegerPoint{Name: "cpu", Time: 0 * Second, Value: 20}},
				{&query.IntegerPoint{Name: "cpu", Time: 0 * Second, Value: 39}},
				{&query.IntegerPoint{Name: "cpu", Time: 4 * Second, Value: 49}},
				{&query.IntegerPoint{Name: "cpu", Time: 4 * Second, Value: 52}},
			},
		},
		{
			name: "HoltWinters_GroupBy_Agg",
			q:    `SELECT holt_winters(mean(value), 2, 2) FROM cpu WHERE time >= '1970-01-01T00:00:10Z' AND time < '1970-01-01T00:00:20Z' GROUP BY time(2s)`,
			typ:  influxql.Float,
			expr: `mean(value::float)`,
			itrs: []query.Iterator{
				&FloatIterator{Points: []query.FloatPoint{
					{Name: "cpu", Time: 10 * Second, Value: 4},
					{Name: "cpu", Time: 11 * Second, Value: 6},

					{Name: "cpu", Time: 12 * Second, Value: 9},
					{Name: "cpu", Time: 13 * Second, Value: 11},

					{Name: "cpu", Time: 14 * Second, Value: 5},
					{Name: "cpu", Time: 15 * Second, Value: 7},

					{Name: "cpu", Time: 16 * Second, Value: 10},
					{Name: "cpu", Time: 17 * Second, Value: 12},

					{Name: "cpu", Time: 18 * Second, Value: 6},
					{Name: "cpu", Time: 19 * Second, Value: 8},
				}},
			},
			points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Time: 20 * Second, Value: 11.960623419918432}},
				{&query.FloatPoint{Name: "cpu", Time: 22 * Second, Value: 7.953140268154609}},
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			shardMapper := ShardMapper{
				MapShardsFn: func(sources influxql.Sources, _ influxql.TimeRange) query.ShardGroup {
					return &ShardGroup{
						Fields: map[string]influxql.DataType{
							"value": tt.typ,
						},
						Dimensions: []string{"host", "region"},
						CreateIteratorFn: func(m *influxql.Measurement, opt query.IteratorOptions) (query.Iterator, error) {
							if m.Name != "cpu" {
								t.Fatalf("unexpected source: %s", m.Name)
							}
							if tt.expr != "" && !reflect.DeepEqual(opt.Expr, MustParseExpr(tt.expr)) {
								t.Fatalf("unexpected expr: %s", spew.Sdump(opt.Expr))
							}

							itrs := tt.itrs
							if _, ok := opt.Expr.(*influxql.Call); ok {
								for i, itr := range itrs {
									itr, err := query.NewCallIterator(itr, opt)
									if err != nil {
										return nil, err
									}
									itrs[i] = itr
								}
							}
							return query.Iterators(itrs).Merge(opt)
						},
					}
				},
			}

			itrs, _, err := query.Select(MustParseSelectStatement(tt.q), &shardMapper, query.SelectOptions{})
			if err != nil {
				if tt.err == "" {
					t.Fatal(err)
				} else if have, want := err.Error(), tt.err; have != want {
					t.Fatalf("unexpected error: have=%s want=%s", have, want)
				}
			} else if tt.err != "" {
				t.Fatal("expected error")
			} else if a, err := Iterators(itrs).ReadAll(); err != nil {
				t.Fatalf("unexpected point: %s", err)
			} else if diff := cmp.Diff(a, tt.points); diff != "" {
				t.Fatalf("unexpected points:\n%s", diff)
			}
		})
	}
}

// Ensure a SELECT binary expr queries can be executed as floats.
func TestSelect_BinaryExpr_Float(t *testing.T) {
	shardMapper := ShardMapper{
		MapShardsFn: func(sources influxql.Sources, _ influxql.TimeRange) query.ShardGroup {
			return &ShardGroup{
				Fields: map[string]influxql.DataType{
					"value": influxql.Float,
				},
				CreateIteratorFn: func(m *influxql.Measurement, opt query.IteratorOptions) (query.Iterator, error) {
					if m.Name != "cpu" {
						t.Fatalf("unexpected source: %s", m.Name)
					}
					makeAuxFields := func(value float64) []interface{} {
						aux := make([]interface{}, len(opt.Aux))
						for i := range aux {
							aux[i] = value
						}
						return aux
					}
					return &FloatIterator{Points: []query.FloatPoint{
						{Name: "cpu", Time: 0 * Second, Aux: makeAuxFields(20)},
						{Name: "cpu", Time: 5 * Second, Aux: makeAuxFields(10)},
						{Name: "cpu", Time: 9 * Second, Aux: makeAuxFields(19)},
					}}, nil
				},
			}
		},
	}

	for _, test := range []struct {
		Name      string
		Statement string
		Points    [][]query.Point
	}{
		{
			Name:      "AdditionRHS_Number",
			Statement: `SELECT value + 2.0 FROM cpu`,
			Points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Time: 0 * Second, Value: 22}},
				{&query.FloatPoint{Name: "cpu", Time: 5 * Second, Value: 12}},
				{&query.FloatPoint{Name: "cpu", Time: 9 * Second, Value: 21}},
			},
		},
		{
			Name:      "AdditionRHS_Integer",
			Statement: `SELECT value + 2 FROM cpu`,
			Points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Time: 0 * Second, Value: 22}},
				{&query.FloatPoint{Name: "cpu", Time: 5 * Second, Value: 12}},
				{&query.FloatPoint{Name: "cpu", Time: 9 * Second, Value: 21}},
			},
		},
		{
			Name:      "AdditionLHS_Number",
			Statement: `SELECT 2.0 + value FROM cpu`,
			Points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Time: 0 * Second, Value: 22}},
				{&query.FloatPoint{Name: "cpu", Time: 5 * Second, Value: 12}},
				{&query.FloatPoint{Name: "cpu", Time: 9 * Second, Value: 21}},
			},
		},
		{
			Name:      "AdditionLHS_Integer",
			Statement: `SELECT 2 + value FROM cpu`,
			Points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Time: 0 * Second, Value: 22}},
				{&query.FloatPoint{Name: "cpu", Time: 5 * Second, Value: 12}},
				{&query.FloatPoint{Name: "cpu", Time: 9 * Second, Value: 21}},
			},
		},
		{
			Name:      "TwoVariableAddition",
			Statement: `SELECT value + value FROM cpu`,
			Points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Time: 0 * Second, Value: 40}},
				{&query.FloatPoint{Name: "cpu", Time: 5 * Second, Value: 20}},
				{&query.FloatPoint{Name: "cpu", Time: 9 * Second, Value: 38}},
			},
		},
		{
			Name:      "MultiplicationRHS_Number",
			Statement: `SELECT value * 2.0 FROM cpu`,
			Points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Time: 0 * Second, Value: 40}},
				{&query.FloatPoint{Name: "cpu", Time: 5 * Second, Value: 20}},
				{&query.FloatPoint{Name: "cpu", Time: 9 * Second, Value: 38}},
			},
		},
		{
			Name:      "MultiplicationRHS_Integer",
			Statement: `SELECT value * 2 FROM cpu`,
			Points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Time: 0 * Second, Value: 40}},
				{&query.FloatPoint{Name: "cpu", Time: 5 * Second, Value: 20}},
				{&query.FloatPoint{Name: "cpu", Time: 9 * Second, Value: 38}},
			},
		},
		{
			Name:      "MultiplicationLHS_Number",
			Statement: `SELECT 2.0 * value FROM cpu`,
			Points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Time: 0 * Second, Value: 40}},
				{&query.FloatPoint{Name: "cpu", Time: 5 * Second, Value: 20}},
				{&query.FloatPoint{Name: "cpu", Time: 9 * Second, Value: 38}},
			},
		},
		{
			Name:      "MultiplicationLHS_Integer",
			Statement: `SELECT 2 * value FROM cpu`,
			Points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Time: 0 * Second, Value: 40}},
				{&query.FloatPoint{Name: "cpu", Time: 5 * Second, Value: 20}},
				{&query.FloatPoint{Name: "cpu", Time: 9 * Second, Value: 38}},
			},
		},
		{
			Name:      "TwoVariableMultiplication",
			Statement: `SELECT value * value FROM cpu`,
			Points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Time: 0 * Second, Value: 400}},
				{&query.FloatPoint{Name: "cpu", Time: 5 * Second, Value: 100}},
				{&query.FloatPoint{Name: "cpu", Time: 9 * Second, Value: 361}},
			},
		},
		{
			Name:      "SubtractionRHS_Number",
			Statement: `SELECT value - 2.0 FROM cpu`,
			Points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Time: 0 * Second, Value: 18}},
				{&query.FloatPoint{Name: "cpu", Time: 5 * Second, Value: 8}},
				{&query.FloatPoint{Name: "cpu", Time: 9 * Second, Value: 17}},
			},
		},
		{
			Name:      "SubtractionRHS_Integer",
			Statement: `SELECT value - 2 FROM cpu`,
			Points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Time: 0 * Second, Value: 18}},
				{&query.FloatPoint{Name: "cpu", Time: 5 * Second, Value: 8}},
				{&query.FloatPoint{Name: "cpu", Time: 9 * Second, Value: 17}},
			},
		},
		{
			Name:      "SubtractionLHS_Number",
			Statement: `SELECT 2.0 - value FROM cpu`,
			Points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Time: 0 * Second, Value: -18}},
				{&query.FloatPoint{Name: "cpu", Time: 5 * Second, Value: -8}},
				{&query.FloatPoint{Name: "cpu", Time: 9 * Second, Value: -17}},
			},
		},
		{
			Name:      "SubtractionLHS_Integer",
			Statement: `SELECT 2 - value FROM cpu`,
			Points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Time: 0 * Second, Value: -18}},
				{&query.FloatPoint{Name: "cpu", Time: 5 * Second, Value: -8}},
				{&query.FloatPoint{Name: "cpu", Time: 9 * Second, Value: -17}},
			},
		},
		{
			Name:      "TwoVariableSubtraction",
			Statement: `SELECT value - value FROM cpu`,
			Points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Time: 0 * Second, Value: 0}},
				{&query.FloatPoint{Name: "cpu", Time: 5 * Second, Value: 0}},
				{&query.FloatPoint{Name: "cpu", Time: 9 * Second, Value: 0}},
			},
		},
		{
			Name:      "DivisionRHS_Number",
			Statement: `SELECT value / 2.0 FROM cpu`,
			Points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Time: 0 * Second, Value: 10}},
				{&query.FloatPoint{Name: "cpu", Time: 5 * Second, Value: 5}},
				{&query.FloatPoint{Name: "cpu", Time: 9 * Second, Value: float64(19) / 2}},
			},
		},
		{
			Name:      "DivisionRHS_Integer",
			Statement: `SELECT value / 2 FROM cpu`,
			Points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Time: 0 * Second, Value: 10}},
				{&query.FloatPoint{Name: "cpu", Time: 5 * Second, Value: 5}},
				{&query.FloatPoint{Name: "cpu", Time: 9 * Second, Value: float64(19) / 2}},
			},
		},
		{
			Name:      "DivisionLHS_Number",
			Statement: `SELECT 38.0 / value FROM cpu`,
			Points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Time: 0 * Second, Value: 1.9}},
				{&query.FloatPoint{Name: "cpu", Time: 5 * Second, Value: 3.8}},
				{&query.FloatPoint{Name: "cpu", Time: 9 * Second, Value: 2}},
			},
		},
		{
			Name:      "DivisionLHS_Integer",
			Statement: `SELECT 38 / value FROM cpu`,
			Points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Time: 0 * Second, Value: 1.9}},
				{&query.FloatPoint{Name: "cpu", Time: 5 * Second, Value: 3.8}},
				{&query.FloatPoint{Name: "cpu", Time: 9 * Second, Value: 2}},
			},
		},
		{
			Name:      "TwoVariableDivision",
			Statement: `SELECT value / value FROM cpu`,
			Points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Time: 0 * Second, Value: 1}},
				{&query.FloatPoint{Name: "cpu", Time: 5 * Second, Value: 1}},
				{&query.FloatPoint{Name: "cpu", Time: 9 * Second, Value: 1}},
			},
		},
	} {
		t.Run(test.Name, func(t *testing.T) {
			stmt := MustParseSelectStatement(test.Statement)
			itrs, _, err := query.Select(stmt, &shardMapper, query.SelectOptions{})
			if err != nil {
				t.Errorf("%s: parse error: %s", test.Name, err)
			} else if a, err := Iterators(itrs).ReadAll(); err != nil {
				t.Fatalf("%s: unexpected error: %s", test.Name, err)
			} else if diff := cmp.Diff(a, test.Points); diff != "" {
				t.Errorf("%s: unexpected points:\n%s", test.Name, diff)
			}
		})
	}
}

// Ensure a SELECT binary expr queries can be executed as integers.
func TestSelect_BinaryExpr_Integer(t *testing.T) {
	shardMapper := ShardMapper{
		MapShardsFn: func(sources influxql.Sources, _ influxql.TimeRange) query.ShardGroup {
			return &ShardGroup{
				Fields: map[string]influxql.DataType{
					"value": influxql.Integer,
				},
				CreateIteratorFn: func(m *influxql.Measurement, opt query.IteratorOptions) (query.Iterator, error) {
					if m.Name != "cpu" {
						t.Fatalf("unexpected source: %s", m.Name)
					}
					makeAuxFields := func(value int64) []interface{} {
						aux := make([]interface{}, len(opt.Aux))
						for i := range aux {
							aux[i] = value
						}
						return aux
					}
					return &FloatIterator{Points: []query.FloatPoint{
						{Name: "cpu", Time: 0 * Second, Aux: makeAuxFields(20)},
						{Name: "cpu", Time: 5 * Second, Aux: makeAuxFields(10)},
						{Name: "cpu", Time: 9 * Second, Aux: makeAuxFields(19)},
					}}, nil
				},
			}
		},
	}

	for _, test := range []struct {
		Name      string
		Statement string
		Points    [][]query.Point
	}{
		{
			Name:      "AdditionRHS_Number",
			Statement: `SELECT value + 2.0 FROM cpu`,
			Points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Time: 0 * Second, Value: 22}},
				{&query.FloatPoint{Name: "cpu", Time: 5 * Second, Value: 12}},
				{&query.FloatPoint{Name: "cpu", Time: 9 * Second, Value: 21}},
			},
		},
		{
			Name:      "AdditionRHS_Integer",
			Statement: `SELECT value + 2 FROM cpu`,
			Points: [][]query.Point{
				{&query.IntegerPoint{Name: "cpu", Time: 0 * Second, Value: 22}},
				{&query.IntegerPoint{Name: "cpu", Time: 5 * Second, Value: 12}},
				{&query.IntegerPoint{Name: "cpu", Time: 9 * Second, Value: 21}},
			},
		},
		{
			Name:      "AdditionLHS_Number",
			Statement: `SELECT 2.0 + value FROM cpu`,
			Points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Time: 0 * Second, Value: 22}},
				{&query.FloatPoint{Name: "cpu", Time: 5 * Second, Value: 12}},
				{&query.FloatPoint{Name: "cpu", Time: 9 * Second, Value: 21}},
			},
		},
		{
			Name:      "AdditionLHS_Integer",
			Statement: `SELECT 2 + value FROM cpu`,
			Points: [][]query.Point{
				{&query.IntegerPoint{Name: "cpu", Time: 0 * Second, Value: 22}},
				{&query.IntegerPoint{Name: "cpu", Time: 5 * Second, Value: 12}},
				{&query.IntegerPoint{Name: "cpu", Time: 9 * Second, Value: 21}},
			},
		},
		{
			Name:      "TwoVariableAddition",
			Statement: `SELECT value + value FROM cpu`,
			Points: [][]query.Point{
				{&query.IntegerPoint{Name: "cpu", Time: 0 * Second, Value: 40}},
				{&query.IntegerPoint{Name: "cpu", Time: 5 * Second, Value: 20}},
				{&query.IntegerPoint{Name: "cpu", Time: 9 * Second, Value: 38}},
			},
		},
		{
			Name:      "MultiplicationRHS_Number",
			Statement: `SELECT value * 2.0 FROM cpu`,
			Points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Time: 0 * Second, Value: 40}},
				{&query.FloatPoint{Name: "cpu", Time: 5 * Second, Value: 20}},
				{&query.FloatPoint{Name: "cpu", Time: 9 * Second, Value: 38}},
			},
		},
		{
			Name:      "MultiplicationRHS_Integer",
			Statement: `SELECT value * 2 FROM cpu`,
			Points: [][]query.Point{
				{&query.IntegerPoint{Name: "cpu", Time: 0 * Second, Value: 40}},
				{&query.IntegerPoint{Name: "cpu", Time: 5 * Second, Value: 20}},
				{&query.IntegerPoint{Name: "cpu", Time: 9 * Second, Value: 38}},
			},
		},
		{
			Name:      "MultiplicationLHS_Number",
			Statement: `SELECT 2.0 * value FROM cpu`,
			Points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Time: 0 * Second, Value: 40}},
				{&query.FloatPoint{Name: "cpu", Time: 5 * Second, Value: 20}},
				{&query.FloatPoint{Name: "cpu", Time: 9 * Second, Value: 38}},
			},
		},
		{
			Name:      "MultiplicationLHS_Integer",
			Statement: `SELECT 2 * value FROM cpu`,
			Points: [][]query.Point{
				{&query.IntegerPoint{Name: "cpu", Time: 0 * Second, Value: 40}},
				{&query.IntegerPoint{Name: "cpu", Time: 5 * Second, Value: 20}},
				{&query.IntegerPoint{Name: "cpu", Time: 9 * Second, Value: 38}},
			},
		},
		{
			Name:      "TwoVariableMultiplication",
			Statement: `SELECT value * value FROM cpu`,
			Points: [][]query.Point{
				{&query.IntegerPoint{Name: "cpu", Time: 0 * Second, Value: 400}},
				{&query.IntegerPoint{Name: "cpu", Time: 5 * Second, Value: 100}},
				{&query.IntegerPoint{Name: "cpu", Time: 9 * Second, Value: 361}},
			},
		},
		{
			Name:      "SubtractionRHS_Number",
			Statement: `SELECT value - 2.0 FROM cpu`,
			Points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Time: 0 * Second, Value: 18}},
				{&query.FloatPoint{Name: "cpu", Time: 5 * Second, Value: 8}},
				{&query.FloatPoint{Name: "cpu", Time: 9 * Second, Value: 17}},
			},
		},
		{
			Name:      "SubtractionRHS_Integer",
			Statement: `SELECT value - 2 FROM cpu`,
			Points: [][]query.Point{
				{&query.IntegerPoint{Name: "cpu", Time: 0 * Second, Value: 18}},
				{&query.IntegerPoint{Name: "cpu", Time: 5 * Second, Value: 8}},
				{&query.IntegerPoint{Name: "cpu", Time: 9 * Second, Value: 17}},
			},
		},
		{
			Name:      "SubtractionLHS_Number",
			Statement: `SELECT 2.0 - value FROM cpu`,
			Points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Time: 0 * Second, Value: -18}},
				{&query.FloatPoint{Name: "cpu", Time: 5 * Second, Value: -8}},
				{&query.FloatPoint{Name: "cpu", Time: 9 * Second, Value: -17}},
			},
		},
		{
			Name:      "SubtractionLHS_Integer",
			Statement: `SELECT 2 - value FROM cpu`,
			Points: [][]query.Point{
				{&query.IntegerPoint{Name: "cpu", Time: 0 * Second, Value: -18}},
				{&query.IntegerPoint{Name: "cpu", Time: 5 * Second, Value: -8}},
				{&query.IntegerPoint{Name: "cpu", Time: 9 * Second, Value: -17}},
			},
		},
		{
			Name:      "TwoVariableSubtraction",
			Statement: `SELECT value - value FROM cpu`,
			Points: [][]query.Point{
				{&query.IntegerPoint{Name: "cpu", Time: 0 * Second, Value: 0}},
				{&query.IntegerPoint{Name: "cpu", Time: 5 * Second, Value: 0}},
				{&query.IntegerPoint{Name: "cpu", Time: 9 * Second, Value: 0}},
			},
		},
		{
			Name:      "DivisionRHS_Number",
			Statement: `SELECT value / 2.0 FROM cpu`,
			Points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Time: 0 * Second, Value: 10}},
				{&query.FloatPoint{Name: "cpu", Time: 5 * Second, Value: 5}},
				{&query.FloatPoint{Name: "cpu", Time: 9 * Second, Value: 9.5}},
			},
		},
		{
			Name:      "DivisionRHS_Integer",
			Statement: `SELECT value / 2 FROM cpu`,
			Points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Time: 0 * Second, Value: 10}},
				{&query.FloatPoint{Name: "cpu", Time: 5 * Second, Value: 5}},
				{&query.FloatPoint{Name: "cpu", Time: 9 * Second, Value: float64(19) / 2}},
			},
		},
		{
			Name:      "DivisionLHS_Number",
			Statement: `SELECT 38.0 / value FROM cpu`,
			Points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Time: 0 * Second, Value: 1.9}},
				{&query.FloatPoint{Name: "cpu", Time: 5 * Second, Value: 3.8}},
				{&query.FloatPoint{Name: "cpu", Time: 9 * Second, Value: 2.0}},
			},
		},
		{
			Name:      "DivisionLHS_Integer",
			Statement: `SELECT 38 / value FROM cpu`,
			Points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Time: 0 * Second, Value: 1.9}},
				{&query.FloatPoint{Name: "cpu", Time: 5 * Second, Value: 3.8}},
				{&query.FloatPoint{Name: "cpu", Time: 9 * Second, Value: 2}},
			},
		},
		{
			Name:      "TwoVariablesDivision",
			Statement: `SELECT value / value FROM cpu`,
			Points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Time: 0 * Second, Value: 1}},
				{&query.FloatPoint{Name: "cpu", Time: 5 * Second, Value: 1}},
				{&query.FloatPoint{Name: "cpu", Time: 9 * Second, Value: 1}},
			},
		},
		{
			Name:      "BitwiseAndRHS",
			Statement: `SELECT value & 254 FROM cpu`,
			Points: [][]query.Point{
				{&query.IntegerPoint{Name: "cpu", Time: 0 * Second, Value: 20}},
				{&query.IntegerPoint{Name: "cpu", Time: 5 * Second, Value: 10}},
				{&query.IntegerPoint{Name: "cpu", Time: 9 * Second, Value: 18}},
			},
		},
		{
			Name:      "BitwiseOrLHS",
			Statement: `SELECT 4 | value FROM cpu`,
			Points: [][]query.Point{
				{&query.IntegerPoint{Name: "cpu", Time: 0 * Second, Value: 20}},
				{&query.IntegerPoint{Name: "cpu", Time: 5 * Second, Value: 14}},
				{&query.IntegerPoint{Name: "cpu", Time: 9 * Second, Value: 23}},
			},
		},
		{
			Name:      "TwoVariableBitwiseXOr",
			Statement: `SELECT value ^ value FROM cpu`,
			Points: [][]query.Point{
				{&query.IntegerPoint{Name: "cpu", Time: 0 * Second, Value: 0}},
				{&query.IntegerPoint{Name: "cpu", Time: 5 * Second, Value: 0}},
				{&query.IntegerPoint{Name: "cpu", Time: 9 * Second, Value: 0}},
			},
		},
	} {
		t.Run(test.Name, func(t *testing.T) {
			stmt := MustParseSelectStatement(test.Statement)
			itrs, _, err := query.Select(stmt, &shardMapper, query.SelectOptions{})
			if err != nil {
				t.Errorf("%s: parse error: %s", test.Name, err)
			} else if a, err := Iterators(itrs).ReadAll(); err != nil {
				t.Fatalf("%s: unexpected error: %s", test.Name, err)
			} else if diff := cmp.Diff(a, test.Points); diff != "" {
				t.Errorf("%s: unexpected points:\n%s", test.Name, diff)
			}
		})
	}
}

// Ensure a SELECT binary expr queries can be executed on mixed iterators.
func TestSelect_BinaryExpr_Mixed(t *testing.T) {
	shardMapper := ShardMapper{
		MapShardsFn: func(sources influxql.Sources, _ influxql.TimeRange) query.ShardGroup {
			return &ShardGroup{
				Fields: map[string]influxql.DataType{
					"total": influxql.Float,
					"value": influxql.Integer,
				},
				CreateIteratorFn: func(m *influxql.Measurement, opt query.IteratorOptions) (query.Iterator, error) {
					if m.Name != "cpu" {
						t.Fatalf("unexpected source: %s", m.Name)
					}
					return &FloatIterator{Points: []query.FloatPoint{
						{Name: "cpu", Time: 0 * Second, Aux: []interface{}{float64(20), int64(10)}},
						{Name: "cpu", Time: 5 * Second, Aux: []interface{}{float64(10), int64(15)}},
						{Name: "cpu", Time: 9 * Second, Aux: []interface{}{float64(19), int64(5)}},
					}}, nil
				},
			}
		},
	}

	for _, test := range []struct {
		Name      string
		Statement string
		Points    [][]query.Point
	}{
		{
			Name:      "Addition",
			Statement: `SELECT total + value FROM cpu`,
			Points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Time: 0 * Second, Value: 30}},
				{&query.FloatPoint{Name: "cpu", Time: 5 * Second, Value: 25}},
				{&query.FloatPoint{Name: "cpu", Time: 9 * Second, Value: 24}},
			},
		},
		{
			Name:      "Subtraction",
			Statement: `SELECT total - value FROM cpu`,
			Points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Time: 0 * Second, Value: 10}},
				{&query.FloatPoint{Name: "cpu", Time: 5 * Second, Value: -5}},
				{&query.FloatPoint{Name: "cpu", Time: 9 * Second, Value: 14}},
			},
		},
		{
			Name:      "Multiplication",
			Statement: `SELECT total * value FROM cpu`,
			Points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Time: 0 * Second, Value: 200}},
				{&query.FloatPoint{Name: "cpu", Time: 5 * Second, Value: 150}},
				{&query.FloatPoint{Name: "cpu", Time: 9 * Second, Value: 95}},
			},
		},
		{
			Name:      "Division",
			Statement: `SELECT total / value FROM cpu`,
			Points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Time: 0 * Second, Value: 2}},
				{&query.FloatPoint{Name: "cpu", Time: 5 * Second, Value: float64(10) / float64(15)}},
				{&query.FloatPoint{Name: "cpu", Time: 9 * Second, Value: float64(19) / float64(5)}},
			},
		},
	} {
		t.Run(test.Name, func(t *testing.T) {
			stmt := MustParseSelectStatement(test.Statement)
			itrs, _, err := query.Select(stmt, &shardMapper, query.SelectOptions{})
			if err != nil {
				t.Errorf("%s: parse error: %s", test.Name, err)
			} else if a, err := Iterators(itrs).ReadAll(); err != nil {
				t.Fatalf("%s: unexpected error: %s", test.Name, err)
			} else if diff := cmp.Diff(a, test.Points); diff != "" {
				t.Errorf("%s: unexpected points:\n%s", test.Name, diff)
			}
		})
	}
}

// Ensure a SELECT binary expr queries can be executed as booleans.
func TestSelect_BinaryExpr_Boolean(t *testing.T) {
	shardMapper := ShardMapper{
		MapShardsFn: func(sources influxql.Sources, _ influxql.TimeRange) query.ShardGroup {
			return &ShardGroup{
				Fields: map[string]influxql.DataType{
					"one": influxql.Boolean,
					"two": influxql.Boolean,
				},
				CreateIteratorFn: func(m *influxql.Measurement, opt query.IteratorOptions) (query.Iterator, error) {
					if m.Name != "cpu" {
						t.Fatalf("unexpected source: %s", m.Name)
					}
					makeAuxFields := func(value bool) []interface{} {
						aux := make([]interface{}, len(opt.Aux))
						for i := range aux {
							aux[i] = value
						}
						return aux
					}
					return &FloatIterator{Points: []query.FloatPoint{
						{Name: "cpu", Time: 0 * Second, Aux: makeAuxFields(true)},
						{Name: "cpu", Time: 5 * Second, Aux: makeAuxFields(false)},
						{Name: "cpu", Time: 9 * Second, Aux: makeAuxFields(true)},
					}}, nil
				},
			}
		},
	}

	for _, test := range []struct {
		Name      string
		Statement string
		Points    [][]query.Point
	}{
		{
			Name:      "BinaryXOrRHS",
			Statement: `SELECT one ^ true FROM cpu`,
			Points: [][]query.Point{
				{&query.BooleanPoint{Name: "cpu", Time: 0 * Second, Value: false}},
				{&query.BooleanPoint{Name: "cpu", Time: 5 * Second, Value: true}},
				{&query.BooleanPoint{Name: "cpu", Time: 9 * Second, Value: false}},
			},
		},
		{
			Name:      "BinaryOrLHS",
			Statement: `SELECT true | two FROM cpu`,
			Points: [][]query.Point{
				{&query.BooleanPoint{Name: "cpu", Time: 0 * Second, Value: true}},
				{&query.BooleanPoint{Name: "cpu", Time: 5 * Second, Value: true}},
				{&query.BooleanPoint{Name: "cpu", Time: 9 * Second, Value: true}},
			},
		},
		{
			Name:      "TwoSeriesBitwiseAnd",
			Statement: `SELECT one & two FROM cpu`,
			Points: [][]query.Point{
				{&query.BooleanPoint{Name: "cpu", Time: 0 * Second, Value: true}},
				{&query.BooleanPoint{Name: "cpu", Time: 5 * Second, Value: false}},
				{&query.BooleanPoint{Name: "cpu", Time: 9 * Second, Value: true}},
			},
		},
	} {
		t.Run(test.Name, func(t *testing.T) {
			stmt := MustParseSelectStatement(test.Statement)
			itrs, _, err := query.Select(stmt, &shardMapper, query.SelectOptions{})
			if err != nil {
				t.Errorf("%s: parse error: %s", test.Name, err)
			} else if a, err := Iterators(itrs).ReadAll(); err != nil {
				t.Fatalf("%s: unexpected error: %s", test.Name, err)
			} else if diff := cmp.Diff(a, test.Points); diff != "" {
				t.Errorf("%s: unexpected points:\n%s", test.Name, diff)
			}
		})
	}
}

// Ensure a SELECT binary expr with nil values can be executed.
// Nil values may be present when a field is missing from one iterator,
// but not the other.
func TestSelect_BinaryExpr_NilValues(t *testing.T) {
	shardMapper := ShardMapper{
		MapShardsFn: func(sources influxql.Sources, _ influxql.TimeRange) query.ShardGroup {
			return &ShardGroup{
				Fields: map[string]influxql.DataType{
					"total": influxql.Float,
					"value": influxql.Float,
				},
				CreateIteratorFn: func(m *influxql.Measurement, opt query.IteratorOptions) (query.Iterator, error) {
					if m.Name != "cpu" {
						t.Fatalf("unexpected source: %s", m.Name)
					}
					return &FloatIterator{Points: []query.FloatPoint{
						{Name: "cpu", Time: 0 * Second, Value: 20, Aux: []interface{}{float64(20), nil}},
						{Name: "cpu", Time: 5 * Second, Value: 10, Aux: []interface{}{float64(10), float64(15)}},
						{Name: "cpu", Time: 9 * Second, Value: 19, Aux: []interface{}{nil, float64(5)}},
					}}, nil
				},
			}
		},
	}

	for _, test := range []struct {
		Name      string
		Statement string
		Points    [][]query.Point
	}{
		{
			Name:      "Addition",
			Statement: `SELECT total + value FROM cpu`,
			Points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Time: 0 * Second, Nil: true}},
				{&query.FloatPoint{Name: "cpu", Time: 5 * Second, Value: 25}},
				{&query.FloatPoint{Name: "cpu", Time: 9 * Second, Nil: true}},
			},
		},
		{
			Name:      "Subtraction",
			Statement: `SELECT total - value FROM cpu`,
			Points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Time: 0 * Second, Nil: true}},
				{&query.FloatPoint{Name: "cpu", Time: 5 * Second, Value: -5}},
				{&query.FloatPoint{Name: "cpu", Time: 9 * Second, Nil: true}},
			},
		},
		{
			Name:      "Multiplication",
			Statement: `SELECT total * value FROM cpu`,
			Points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Time: 0 * Second, Nil: true}},
				{&query.FloatPoint{Name: "cpu", Time: 5 * Second, Value: 150}},
				{&query.FloatPoint{Name: "cpu", Time: 9 * Second, Nil: true}},
			},
		},
		{
			Name:      "Division",
			Statement: `SELECT total / value FROM cpu`,
			Points: [][]query.Point{
				{&query.FloatPoint{Name: "cpu", Time: 0 * Second, Nil: true}},
				{&query.FloatPoint{Name: "cpu", Time: 5 * Second, Value: float64(10) / float64(15)}},
				{&query.FloatPoint{Name: "cpu", Time: 9 * Second, Nil: true}},
			},
		},
	} {
		t.Run(test.Name, func(t *testing.T) {
			stmt := MustParseSelectStatement(test.Statement)
			itrs, _, err := query.Select(stmt, &shardMapper, query.SelectOptions{})
			if err != nil {
				t.Errorf("%s: parse error: %s", test.Name, err)
			} else if a, err := Iterators(itrs).ReadAll(); err != nil {
				t.Fatalf("%s: unexpected error: %s", test.Name, err)
			} else if diff := cmp.Diff(a, test.Points); diff != "" {
				t.Errorf("%s: unexpected points:\n%s", test.Name, diff)
			}
		})
	}
}

type ShardMapper struct {
	MapShardsFn func(sources influxql.Sources, t influxql.TimeRange) query.ShardGroup
}

func (m *ShardMapper) MapShards(sources influxql.Sources, t influxql.TimeRange, opt query.SelectOptions) (query.ShardGroup, error) {
	shards := m.MapShardsFn(sources, t)
	return shards, nil
}

type ShardGroup struct {
	CreateIteratorFn func(m *influxql.Measurement, opt query.IteratorOptions) (query.Iterator, error)
	Fields           map[string]influxql.DataType
	Dimensions       []string
}

func (sh *ShardGroup) CreateIterator(m *influxql.Measurement, opt query.IteratorOptions) (query.Iterator, error) {
	return sh.CreateIteratorFn(m, opt)
}

func (sh *ShardGroup) IteratorCost(m *influxql.Measurement, opt query.IteratorOptions) (query.IteratorCost, error) {
	return query.IteratorCost{}, nil
}

func (sh *ShardGroup) FieldDimensions(m *influxql.Measurement) (fields map[string]influxql.DataType, dimensions map[string]struct{}, err error) {
	fields = make(map[string]influxql.DataType)
	dimensions = make(map[string]struct{})

	for f, typ := range sh.Fields {
		fields[f] = typ
	}
	for _, d := range sh.Dimensions {
		dimensions[d] = struct{}{}
	}
	return fields, dimensions, nil
}

func (sh *ShardGroup) MapType(m *influxql.Measurement, field string) influxql.DataType {
	if typ, ok := sh.Fields[field]; ok {
		return typ
	}
	for _, d := range sh.Dimensions {
		if d == field {
			return influxql.Tag
		}
	}
	return influxql.Unknown
}

func (*ShardGroup) Close() error {
	return nil
}

func BenchmarkSelect_Raw_1K(b *testing.B)   { benchmarkSelectRaw(b, 1000) }
func BenchmarkSelect_Raw_100K(b *testing.B) { benchmarkSelectRaw(b, 1000000) }

func benchmarkSelectRaw(b *testing.B, pointN int) {
	benchmarkSelect(b, MustParseSelectStatement(`SELECT fval FROM cpu`), NewRawBenchmarkIteratorCreator(pointN))
}

func benchmarkSelect(b *testing.B, stmt *influxql.SelectStatement, shardMapper query.ShardMapper) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		itrs, _, err := query.Select(stmt, shardMapper, query.SelectOptions{})
		if err != nil {
			b.Fatal(err)
		}
		query.DrainIterators(itrs)
	}
}

// NewRawBenchmarkIteratorCreator returns a new mock iterator creator with generated fields.
func NewRawBenchmarkIteratorCreator(pointN int) query.ShardMapper {
	return &ShardMapper{
		MapShardsFn: func(sources influxql.Sources, t influxql.TimeRange) query.ShardGroup {
			return &ShardGroup{
				Fields: map[string]influxql.DataType{
					"fval": influxql.Float,
				},
				CreateIteratorFn: func(m *influxql.Measurement, opt query.IteratorOptions) (query.Iterator, error) {
					if opt.Expr != nil {
						panic("unexpected expression")
					}

					p := query.FloatPoint{
						Name: "cpu",
						Aux:  make([]interface{}, len(opt.Aux)),
					}

					for i := range opt.Aux {
						switch opt.Aux[i].Val {
						case "fval":
							p.Aux[i] = float64(100)
						default:
							panic("unknown iterator expr: " + opt.Expr.String())
						}
					}

					return &FloatPointGenerator{N: pointN, Fn: func(i int) *query.FloatPoint {
						p.Time = int64(time.Duration(i) * (10 * time.Second))
						return &p
					}}, nil
				},
			}
		},
	}
}

func benchmarkSelectDedupe(b *testing.B, seriesN, pointsPerSeries int) {
	stmt := MustParseSelectStatement(`SELECT sval::string FROM cpu`)
	stmt.Dedupe = true

	shardMapper := ShardMapper{
		MapShardsFn: func(sources influxql.Sources, t influxql.TimeRange) query.ShardGroup {
			return &ShardGroup{
				Fields: map[string]influxql.DataType{
					"sval": influxql.String,
				},
				CreateIteratorFn: func(m *influxql.Measurement, opt query.IteratorOptions) (query.Iterator, error) {
					if opt.Expr != nil {
						panic("unexpected expression")
					}

					p := query.FloatPoint{
						Name: "tags",
						Aux:  []interface{}{nil},
					}

					return &FloatPointGenerator{N: seriesN * pointsPerSeries, Fn: func(i int) *query.FloatPoint {
						p.Aux[0] = fmt.Sprintf("server%d", i%seriesN)
						return &p
					}}, nil
				},
			}
		},
	}

	b.ResetTimer()
	benchmarkSelect(b, stmt, &shardMapper)
}

func BenchmarkSelect_Dedupe_1K(b *testing.B) { benchmarkSelectDedupe(b, 1000, 100) }

func benchmarkSelectTop(b *testing.B, seriesN, pointsPerSeries int) {
	stmt := MustParseSelectStatement(`SELECT top(sval, 10) FROM cpu`)

	shardMapper := ShardMapper{
		MapShardsFn: func(sources influxql.Sources, t influxql.TimeRange) query.ShardGroup {
			return &ShardGroup{
				Fields: map[string]influxql.DataType{
					"sval": influxql.Float,
				},
				CreateIteratorFn: func(m *influxql.Measurement, opt query.IteratorOptions) (query.Iterator, error) {
					if m.Name != "cpu" {
						b.Fatalf("unexpected source: %s", m.Name)
					}
					if !reflect.DeepEqual(opt.Expr, MustParseExpr(`sval`)) {
						b.Fatalf("unexpected expr: %s", spew.Sdump(opt.Expr))
					}

					p := query.FloatPoint{
						Name: "cpu",
					}

					return &FloatPointGenerator{N: seriesN * pointsPerSeries, Fn: func(i int) *query.FloatPoint {
						p.Value = float64(rand.Int63())
						p.Time = int64(time.Duration(i) * (10 * time.Second))
						return &p
					}}, nil
				},
			}
		},
	}

	b.ResetTimer()
	benchmarkSelect(b, stmt, &shardMapper)
}

func BenchmarkSelect_Top_1K(b *testing.B) { benchmarkSelectTop(b, 1000, 1000) }
