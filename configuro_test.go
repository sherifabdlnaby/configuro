package configuro_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sherifabdlnaby/configuro"
)

type Example struct {
	Nested Nested
}

type Nested struct {
	Key   Key
	Key_A Key
	Key_X Key `config:"key-b"`
}

type Key struct {
	A     string
	B     string
	EMPTY string
}

func TestEnvVarsRenaming(t *testing.T) {
	// Set osEnvVars
	_ = os.Setenv("NESTED_KEY_A", "A")
	_ = os.Setenv("NESTED_KEY_B", "B")
	_ = os.Setenv("NESTED_KEY__A_A", "AA")
	_ = os.Setenv("NESTED_KEY__A_B", "AB")
	_ = os.Setenv("NESTED_KEY__B_A", "BA")
	_ = os.Setenv("NESTED_KEY__B_B", "BB")

	envOnlyWithoutPrefix, err := configuro.NewConfigx(
		configuro.LoadFromEnvironmentVariables(true, ""),
		configuro.LoadDotEnvFile(false, ""),
		configuro.LoadFromConfigFile(false, "", "", false, ""),
	)
	if err != nil {
		t.Error(err)
	}

	tests := []struct {
		name     string
		config   *configuro.Config
		expected Example
	}{
		{name: "Renaming", config: envOnlyWithoutPrefix, expected: Example{
			Nested: Nested{
				Key: Key{
					A: "A",
					B: "B",
				},
				Key_A: Key{
					A: "AA",
					B: "AB",
				},
				Key_X: Key{
					A: "BA",
					B: "BB",
				},
			},
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			example := &Example{}
			err := test.config.Load(example)
			if err != nil {
				t.Error(err)
			}

			if example.Nested.Key.A != test.expected.Nested.Key.A ||
				example.Nested.Key.B != test.expected.Nested.Key.B ||
				example.Nested.Key_A.A != test.expected.Nested.Key_A.A ||
				example.Nested.Key_A.B != test.expected.Nested.Key_A.B ||
				example.Nested.Key_X.A != test.expected.Nested.Key_X.A ||
				example.Nested.Key_X.B != test.expected.Nested.Key_X.B {
				t.Errorf("Loaded Values doesn't equal expected values. loaded: %v, expected: %v", example, test.expected)
			}
		})
	}
}

func TestLoadFromEnvVarsOnly(t *testing.T) {
	// Set osEnvVars
	_ = os.Setenv("PREFIX_NESTED_KEY_A", "X")
	_ = os.Setenv("PREFIX_NESTED_KEY_B", "Y")
	_ = os.Setenv("NESTED_KEY_A", "A")
	_ = os.Setenv("NESTED_KEY_B", "B")
	_ = os.Setenv("NESTED_KEY_EMPTY", "")

	envOnlyWithoutPrefix, err := configuro.NewConfigx(
		configuro.LoadFromEnvironmentVariables(true, ""),
		configuro.LoadDotEnvFile(false, ""),
		configuro.LoadFromConfigFile(false, "", "", false, ""),
	)
	if err != nil {
		t.Error(err)
	}

	envOnlyWithPrefix, err := configuro.NewConfigx(
		configuro.LoadFromEnvironmentVariables(true, "PREFIX"),
		configuro.LoadDotEnvFile(false, ""),
		configuro.LoadFromConfigFile(false, "", "", false, ""),
	)
	if err != nil {
		t.Error(err)
	}

	tests := []struct {
		name     string
		config   *configuro.Config
		expected Example
	}{
		{name: "withoutPrefix", config: envOnlyWithoutPrefix, expected: Example{Nested{Key: Key{
			A:     "A",
			B:     "B",
			EMPTY: "",
		}}}},
		{name: "withPrefix", config: envOnlyWithPrefix, expected: Example{Nested{Key: Key{
			A:     "X",
			B:     "Y",
			EMPTY: "",
		}}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			example := &Example{}
			err := test.config.Load(example)
			if err != nil {
				t.Error(err)
			}

			if example.Nested.Key.A != test.expected.Nested.Key.A ||
				example.Nested.Key.B != test.expected.Nested.Key.B ||
				example.Nested.Key.EMPTY != test.expected.Nested.Key.EMPTY {
				t.Error("Loaded Values doesn't equal expected values.")
			}
		})
	}
}

func TestLoadFromFileOnly(t *testing.T) {
	configFileYaml, err := ioutil.TempFile("", "*.yml")
	defer func() {
		configFileYaml.Close()
		os.RemoveAll(configFileYaml.Name())
	}()

	// Write Config to File
	configFileYaml.WriteString(`
nested:
    key:
        a: A
        b: B
    key_a:
        a: AA
        b: AB
    key-b:
        a: BA
        b: BB
    `)

	envOnlyWithoutPrefix, err := configuro.NewConfigx(
		configuro.LoadFromEnvironmentVariables(false, ""),
		configuro.LoadDotEnvFile(false, ""),
		configuro.LoadFromConfigFile(true, strings.TrimSuffix(filepath.Base(configFileYaml.Name()), filepath.Ext(filepath.Base(configFileYaml.Name()))), filepath.Dir(configFileYaml.Name()), false, ""),
	)
	if err != nil {
		t.Error(err)
	}

	tests := []struct {
		name     string
		config   *configuro.Config
		expected Example
	}{
		{name: "withoutPrefix", config: envOnlyWithoutPrefix, expected: Example{
			Nested: Nested{
				Key: Key{
					A: "A",
					B: "B",
				},
				Key_A: Key{
					A: "AA",
					B: "AB",
				},
				Key_X: Key{
					A: "BA",
					B: "BB",
				},
			},
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			example := &Example{}
			err := test.config.Load(example)
			if err != nil {
				t.Error(err)
			}

			if example.Nested.Key.A != test.expected.Nested.Key.A ||
				example.Nested.Key.B != test.expected.Nested.Key.B ||
				example.Nested.Key_A.A != test.expected.Nested.Key_A.A ||
				example.Nested.Key_A.B != test.expected.Nested.Key_A.B ||
				example.Nested.Key_X.A != test.expected.Nested.Key_X.A ||
				example.Nested.Key_X.B != test.expected.Nested.Key_X.B {
				t.Errorf("Loaded Values doesn't equal expected values. loaded: %v, expected: %v", example, test.expected)
			}
		})
	}
}

func TestChangeTagName(t *testing.T) {
	configFileYaml, err := ioutil.TempFile("", "*.yml")
	defer func() {
		configFileYaml.Close()
		os.RemoveAll(configFileYaml.Name())
	}()

	// Write Config to File
	configFileYaml.WriteString(`
Object:
    keyA: aa
    keyB: bb
    key_b: xx
    `)

	type Object struct {
		KeyA string `config:"key_b"`
		KeyB string `newtag:"key_b"`
	}

	type Obj struct {
		Object Object
	}

	configLoaderDefaultTag, err := configuro.NewConfigx(
		configuro.LoadFromEnvironmentVariables(false, ""),
		configuro.LoadDotEnvFile(false, ""),
		configuro.LoadFromConfigFile(true, strings.TrimSuffix(filepath.Base(configFileYaml.Name()), filepath.Ext(filepath.Base(configFileYaml.Name()))), filepath.Dir(configFileYaml.Name()), false, ""),
	)
	if err != nil {
		t.Error(err)
	}

	configLoaderNewTag, err := configuro.NewConfigx(
		configuro.Tag("newtag"),
		configuro.LoadFromEnvironmentVariables(false, ""),
		configuro.LoadDotEnvFile(false, ""),
		configuro.LoadFromConfigFile(true, strings.TrimSuffix(filepath.Base(configFileYaml.Name()), filepath.Ext(filepath.Base(configFileYaml.Name()))), filepath.Dir(configFileYaml.Name()), false, ""),
	)
	if err != nil {
		t.Error(err)
	}

	tests := []struct {
		name     string
		config   *configuro.Config
		expected Obj
	}{
		{name: "defaultTag", config: configLoaderDefaultTag, expected: Obj{
			Object: Object{
				KeyA: "xx",
				KeyB: "bb",
			},
		}},
		{name: "newTag", config: configLoaderNewTag, expected: Obj{
			Object: Object{
				KeyA: "aa",
				KeyB: "xx",
			},
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			example := &Obj{}
			err := test.config.Load(example)
			if err != nil {
				t.Error(err)
			}

			if example.Object.KeyA != test.expected.Object.KeyA ||
				example.Object.KeyB != test.expected.Object.KeyB {
				t.Errorf("Loaded Values doesn't equal expected values. loaded: %v, expected: %v", example, test.expected)
			}
		})
	}
}
