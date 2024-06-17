package validation

import "testing"

func TestKeysInValidator_Validate(t *testing.T) {
	type fields struct {
		BaseValidator BaseValidator
		Keys          []string
	}
	type args struct {
		ctx *Context
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "Valid",
			fields: fields{
				Keys: []string{"key1", "key2"},
			},
			args: args{
				ctx: &Context{
					Value: map[string]interface{}{
						"key1": "value",
						"key2": "value",
					},
				},
			},
			want: true,
		},
		{
			name: "Invalid",
			fields: fields{
				Keys: []string{"key1", "key2"},
			},
			args: args{
				ctx: &Context{
					Value: map[string]interface{}{
						"key1": "value",
					},
				},
			},
			want: false,
		},
		{
			name: "Invalid - Not a map",
			fields: fields{
				Keys: []string{"key1", "key2"},
			},
			args: args{
				ctx: &Context{
					Value: "not a map",
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &KeysInValidator{
				BaseValidator: tt.fields.BaseValidator,
				Keys:          tt.fields.Keys,
			}
			if got := v.Validate(tt.args.ctx); got != tt.want {
				t.Errorf("KeysInValidator.Validate() = %v, want %v", got, tt.want)
			}
		})
	}
}
