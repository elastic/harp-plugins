package linter

import (
	"context"
	"os"
	"reflect"
	"testing"

	linterv1 "github.com/elastic/harp-plugins/cmd/harp-linter/api/gen/go/harp/linter/v1"
	bundlev1 "github.com/elastic/harp/api/gen/go/harp/bundle/v1"
)

func TestValidate(t *testing.T) {
	type args struct {
		spec *linterv1.RuleSet
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "nil",
			wantErr: true,
		},
		{
			name: "invalid apiVersion",
			args: args{
				spec: &linterv1.RuleSet{
					ApiVersion: "foo",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid kind",
			args: args{
				spec: &linterv1.RuleSet{
					ApiVersion: "harp.elastic.co/linter/v1",
					Kind:       "foo",
				},
			},
			wantErr: true,
		},
		{
			name: "nil meta",
			args: args{
				spec: &linterv1.RuleSet{
					ApiVersion: "harp.elastic.co/linter/v1",
					Kind:       "RuleSet",
				},
			},
			wantErr: true,
		},
		{
			name: "meta name not defined",
			args: args{
				spec: &linterv1.RuleSet{
					ApiVersion: "harp.elastic.co/linter/v1",
					Kind:       "RuleSet",
					Meta:       &linterv1.RuleSetMeta{},
				},
			},
			wantErr: true,
		},
		{
			name: "nil spec",
			args: args{
				spec: &linterv1.RuleSet{
					ApiVersion: "harp.elastic.co/linter/v1",
					Kind:       "RuleSet",
					Meta:       &linterv1.RuleSetMeta{},
				},
			},
			wantErr: true,
		},
		{
			name: "no action patch",
			args: args{
				spec: &linterv1.RuleSet{
					ApiVersion: "harp.elastic.co/linter/v1",
					Kind:       "RuleSet",
					Meta:       &linterv1.RuleSetMeta{},
					Spec:       &linterv1.RuleSetSpec{},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Validate(tt.args.spec); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestChecksum(t *testing.T) {
	type args struct {
		spec *linterv1.RuleSet
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "nil",
			wantErr: true,
		},
		{
			name: "valid",
			args: args{
				spec: &linterv1.RuleSet{
					ApiVersion: "harp.elastic.co/linter/v1",
					Kind:       "RuleSet",
					Meta:       &linterv1.RuleSetMeta{},
					Spec:       &linterv1.RuleSetSpec{},
				},
			},
			wantErr: false,
			want:    "j2My0Uf18TvYNBhchM4MnlSm-30RWBhxj7P7QHarZ70",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Checksum(tt.args.spec)
			if (err != nil) != tt.wantErr {
				t.Errorf("Checksum() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Checksum() = %v, want %v", got, tt.want)
			}
		})
	}
}

func mustLoadRuleSet(filePath string) *linterv1.RuleSet {
	f, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}

	p, err := YAML(f)
	if err != nil {
		panic(err)
	}

	return p
}

func TestEvaluate(t *testing.T) {
	type args struct {
		spec *linterv1.RuleSet
		b    *bundlev1.Bundle
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "nil",
			wantErr: true,
		},
		{
			name: "empty bundle",
			args: args{
				spec: mustLoadRuleSet("../../../test/fixtures/ruleset/valid/cso.yaml"),
				b:    &bundlev1.Bundle{},
			},
			wantErr: true,
		},
		{
			name: "cso - invalid bundle",
			args: args{
				spec: mustLoadRuleSet("../../../test/fixtures/ruleset/valid/cso.yaml"),
				b: &bundlev1.Bundle{
					Packages: []*bundlev1.Package{
						{
							Name: "app/qa/security",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "cso - valid bundle",
			args: args{
				spec: mustLoadRuleSet("../../../test/fixtures/ruleset/valid/cso.yaml"),
				b: &bundlev1.Bundle{
					Packages: []*bundlev1.Package{
						{
							Name: "app/qa/security/harp/v1.0.0/server/database/credentials",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "db - valid bundle",
			args: args{
				spec: mustLoadRuleSet("../../../test/fixtures/ruleset/valid/database-secret-validator.yaml"),
				b: &bundlev1.Bundle{
					Packages: []*bundlev1.Package{
						{
							Name: "app/qa/security/harp/v1.0.0/server/database/credentials",
							Secrets: &bundlev1.SecretChain{
								Data: []*bundlev1.KV{
									{
										Key: "DB_HOST",
									},
									{
										Key: "DB_NAME",
									},
									{
										Key: "DB_USER",
									},
									{
										Key: "DB_PASSWORD",
									},
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "db - invalid bundle",
			args: args{
				spec: mustLoadRuleSet("../../../test/fixtures/ruleset/valid/database-secret-validator.yaml"),
				b: &bundlev1.Bundle{
					Packages: []*bundlev1.Package{
						{
							Name: `app/qa/security/harp/v1.0.0/server/database/credentials`,
							Secrets: &bundlev1.SecretChain{
								Data: []*bundlev1.KV{
									{
										Key: "DB_HOST",
									},
								},
							},
						},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Evaluate(context.Background(), tt.args.b, tt.args.spec)
			if (err != nil) != tt.wantErr {
				t.Errorf("Evaluate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
