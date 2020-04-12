package helper

import (
	"encoding/json"
	"testing"

	"github.com/modern-go/parse/model"
	"github.com/stretchr/testify/require"
)

func TestMarshalModelMap(t *testing.T) {
	var testCase = []struct {
		data   model.Map
		expect map[string]interface{}
	}{
		{
			data: model.Map{
				1:   1,
				"2": 2,
				33:  3,
			},
			expect: map[string]interface{}{
				"1":  1,
				"2":  2,
				"33": 3,
			},
		},
		{
			data: model.Map{
				1:   1,
				"2": 2,
				"hello": model.Map{
					"foo": "bar",
				},
			},
			expect: map[string]interface{}{
				"1": 1,
				"2": 2,
				"hello": map[string]interface{}{
					"foo": "bar",
				},
			},
		},
		{
			data: model.Map{
				1:   1,
				"2": 2,
				"hello": model.Map{
					"foo": "bar",
					"baz": map[int]string{
						5:   "hehe",
						100: "one hundread",
					},
				},
			},
			expect: map[string]interface{}{
				"1": 1,
				"2": 2,
				"hello": map[string]interface{}{
					"foo": "bar",
					"baz": map[string]interface{}{
						"5":   "hehe",
						"100": "one hundread",
					},
				},
			},
		},
		{
			data: model.Map{
				"field_1": model.Map{
					"field_2": "4C351760F4A9486E060053B9167FD7FC",
					"field_3": model.Map{
						"val_type": "struct",
						"data": model.List{
							model.Map{
								"field_1": 113.87793,
								"field_2": 23.1099,
							},
						},
					},
					"field_4": 2,
					"field_5": model.Map{
						"field_1": 2,
					},
					"field_6": model.Map{
						"field_1": 128,
					},
					"field_1": 176590671400,
				},
			},
			expect: map[string]interface{}{
				"field_1": map[string]interface{}{
					"field_2": "4C351760F4A9486E060053B9167FD7FC",
					"field_3": map[string]interface{}{
						"val_type": "struct",
						"data": []interface{}{
							map[string]interface{}{
								"field_1": 113.87793,
								"field_2": 23.1099,
							},
						},
					},
					"field_4": 2,
					"field_5": map[string]interface{}{
						"field_1": 2,
					},
					"field_6": map[string]interface{}{
						"field_1": 128,
					},

					"field_1": 176590671400,
				},
			},
		},
	}
	should := require.New(t)
	for idx, tc := range testCase {
		actual, err := MarshalMap(tc.data)
		should.NoError(err)
		expect, err := json.Marshal(tc.expect)
		should.NoError(err)
		should.Equal(expect, actual, "case #%d fail", idx)
	}
}
