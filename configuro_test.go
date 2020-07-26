//nolint
package configuro_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sherifabdlnaby/configuro"
	"go.uber.org/multierr"
)

type Example struct {
	Nested Nested
}

type Nested struct {
	Key         Key
	Key_A       *Key
	Key_X       *Key `config:"key-b"`
	Number      int
	NumberList1 []int
	NumberList2 []int
	KeyList     []Key
	KeyMap      map[string]Key
	KeySlice    []Key
	IntMap      map[string]int
	private     string
}

type Key struct {
	A     string
	B     string
	C     string `validate:"required"`
	D     string `validate:"required"`
	E     string
	EMPTY string
}

type test struct {
	name     string
	config   *configuro.Config
	expected Example
}

func (k Key) Validate() error {
	if k.A != k.B {
		return fmt.Errorf("failed to validate key because A(%s) != B(%s)", k.A, k.B)
	}
	return nil
}

func TestEnvVarsEscaping(t *testing.T) {

	envOnlyWithoutPrefix, err := configuro.NewConfig(
		configuro.WithLoadFromEnvVars("CONFIG"),
		configuro.WithoutLoadDotEnv(),
		configuro.WithoutLoadFromConfigFile(),
	)
	if err != nil {
		t.Fatal(err)
	}

	// Set Values
	_ = os.Setenv("CONFIG_NESTED_KEY_A", "A")
	_ = os.Setenv("CONFIG_NESTED_KEY_B", "B")
	_ = os.Setenv("CONFIG_NESTED_KEY__A_A", "AA")
	_ = os.Setenv("CONFIG_NESTED_KEY__A_B", "AB")
	_ = os.Setenv("CONFIG_NESTED_KEY-B_A", "BA")
	_ = os.Setenv("CONFIG_NESTED_KEY-B_B", "BB")
	_ = os.Setenv("CONFIG_NESTED_KEY-B_B", "BB")
	_ = os.Setenv("CONFIG_NESTED_KEY-B_B", "BB")

	tests := []test{
		{name: "Renaming", config: envOnlyWithoutPrefix, expected: Example{
			Nested: Nested{
				Key: Key{
					A: "A",
					B: "B",
				},
				Key_A: &Key{
					A: "AA",
					B: "AB",
				},
				Key_X: &Key{
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
				t.Fatal(err)
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

	envOnlyWithDefaultPrefix, err := configuro.NewConfig(
		configuro.WithoutLoadDotEnv(),
		configuro.WithoutLoadFromConfigFile(),
		configuro.WithoutEnvConfigPathOverload(),
	)
	if err != nil {
		t.Fatal(err)
	}

	envOnlyWithPrefix, err := configuro.NewConfig(
		configuro.WithLoadFromEnvVars("PREFIX"),
		configuro.WithoutLoadDotEnv(),
		configuro.WithoutLoadFromConfigFile(),
		configuro.WithoutEnvConfigPathOverload(),
		configuro.WithoutExpandEnvVars(),
	)
	if err != nil {
		t.Fatal(err)
	}

	// Set osEnvVars
	_ = os.Setenv("PREFIX_NESTED_KEY_A", "X")
	_ = os.Setenv("PREFIX_NESTED_KEY_B", "Y")
	_ = os.Setenv("CONFIG_NESTED_KEY_A", "A")
	_ = os.Setenv("CONFIG_NESTED_KEY_B", "B")
	_ = os.Setenv("CONFIG_NESTED_KEY_EMPTY", "")

	tests := []test{
		{name: "withoutPrefix", config: envOnlyWithDefaultPrefix, expected: Example{Nested{Key: Key{
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
				t.Fatal(err)
			}
			if example.Nested.Key.A != test.expected.Nested.Key.A ||
				example.Nested.Key.B != test.expected.Nested.Key.B ||
				example.Nested.Key.EMPTY != test.expected.Nested.Key.EMPTY {
				t.Fatal("Loaded Values doesn't equal expected values.")
			}
		})
	}
}

func TestLoadDotEnv(t *testing.T) {

	// Clear Env that may be set up by previous tests. So that .env values are not overridden
	os.Clearenv()

	dotEnvFile, err := ioutil.TempFile("", "*.env")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		dotEnvFile.Close()
		os.RemoveAll(dotEnvFile.Name())
	}()

	// Write Config to File
	dotEnvFile.WriteString(`
PREFIX_NESTED_KEY_A: XXX
PREFIX_NESTED_KEY_B: YYY
#PREFIX_NESTED_KEY_B: XYZ
PREFIX_NESTED_EMPTY:
#PREFIX_NESTED_EMPTY: FD
CONFIG_NESTED_KEY_A: A
CONFIG_NESTED_KEY_B: B
CONFIG_NESTED_KEY_EMPTY:
    `)

	envOnlyWithPrefix, err := configuro.NewConfig(
		configuro.WithLoadFromEnvVars("PREFIX"),
		configuro.WithLoadDotEnv(dotEnvFile.Name()),
		configuro.WithoutLoadFromConfigFile(),
		configuro.WithoutEnvConfigPathOverload(),
	)
	if err != nil {
		t.Fatal(err)
	}

	tests := []test{
		{name: "withPrefix", config: envOnlyWithPrefix, expected: Example{Nested{Key: Key{
			A:     "XXX",
			B:     "YYY",
			EMPTY: "",
		}}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			example := &Example{}
			err := test.config.Load(example)
			if err != nil {
				t.Fatal(err)
			}

			if example.Nested.Key.A != test.expected.Nested.Key.A ||
				example.Nested.Key.B != test.expected.Nested.Key.B ||
				example.Nested.Key.EMPTY != test.expected.Nested.Key.EMPTY {
				t.Fatal("Loaded Values doesn't equal expected values.")
			}
		})
	}
}

func TestLoadFromFileThatDoesntExist(t *testing.T) {
	configLoader, err := configuro.NewConfig(
		configuro.WithLoadFromEnvVars("XXX"),
		configuro.WithLoadDotEnv(""),
		configuro.WithLoadFromConfigFile("filethatdoesntexist.yml", false),
		configuro.WithEnvConfigPathOverload(""),
	)
	if err != nil {
		t.Fatal(err)
	}

	tests := []test{
		{name: "LoadFromFileThatDoesntExist", config: configLoader, expected: Example{}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			example := &Example{}
			err := test.config.Load(example)
			if err != nil {
				t.Fatal(err)
			}

			if example.Nested.Key.A != test.expected.Nested.Key.A ||
				example.Nested.Key.B != test.expected.Nested.Key.B ||
				example.Nested.Key_A != test.expected.Nested.Key_A ||
				example.Nested.Key_X != test.expected.Nested.Key_X {
				t.Fatalf("Loaded Values doesn't equal expected values. loaded: %v, expected: %v", example, test.expected)
			}
		})
	}
}

func TestLoadFromFileThatDoesntExistError(t *testing.T) {
	configLoader, err := configuro.NewConfig(
		configuro.WithLoadFromConfigFile("filethatdoesntexist.yml", true),
	)
	if err != nil {
		t.Fatal(err)
	}

	err = configLoader.Load(&Example{})

	if err == nil {
		t.Fatal("Load should raise error if ErrIfFileNotFound was set to true")
	}
	if !strings.Contains(err.Error(), "error config file not found") {
		t.Fatal("Load should raise not found error if ErrIfFileNotFound was set to true")
	}
}

func TestLoadFromFileOnly(t *testing.T) {
	configFileYaml, err := ioutil.TempFile("", "TestLoadFromFileOnly*.yml")
	if err != nil {
		t.Fatal(err)
	}

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

	configLoader, err := configuro.NewConfig(
		configuro.WithLoadFromEnvVars("X"),
		configuro.WithLoadDotEnv(""),
		configuro.WithLoadFromConfigFile(configFileYaml.Name(), false),
		configuro.WithEnvConfigPathOverload(""),
	)
	if err != nil {
		t.Fatal(err)
	}

	tests := []test{
		{name: "LoadFromFile", config: configLoader, expected: Example{
			Nested: Nested{
				Key: Key{
					A: "A",
					B: "B",
				},
				Key_A: &Key{
					A: "AA",
					B: "AB",
				},
				Key_X: &Key{
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
				t.Fatal(err)
			}

			if example.Nested.Key.A != test.expected.Nested.Key.A ||
				example.Nested.Key.B != test.expected.Nested.Key.B ||
				example.Nested.Key_A.A != test.expected.Nested.Key_A.A ||
				example.Nested.Key_A.B != test.expected.Nested.Key_A.B ||
				example.Nested.Key_X.A != test.expected.Nested.Key_X.A ||
				example.Nested.Key_X.B != test.expected.Nested.Key_X.B {
				t.Fatalf("Loaded Values doesn't equal expected values. loaded: %v, expected: %v", example, test.expected)
			}
		})
	}
}

//TODO Too long, try make it better.
func TestOverloadConfigDirWithEnv(t *testing.T) {

	os.Clearenv()

	err := os.MkdirAll(os.TempDir()+"/conf/", 0777)
	if err != nil {
		t.Fatal(err)
	}
	err = os.MkdirAll(os.TempDir()+"/confOverloaded1/", 0777)
	if err != nil {
		t.Fatal(err)
	}

	err = os.MkdirAll(os.TempDir()+"/confOverloaded2/", 0777)
	if err != nil {
		t.Fatal(err)
	}

	configFileYaml, err := ioutil.TempFile(os.TempDir()+"/conf/", "TestOverloadConfigDirWithEnv*.yml")
	if err != nil {
		t.Fatal(err)
	}

	configFileOverloaded1, err := os.OpenFile(os.TempDir()+"/confOverloaded1/"+filepath.Base(configFileYaml.Name()), os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		t.Fatal(err)
	}

	configFileOverloaded2, err := os.OpenFile(os.TempDir()+"/confOverloaded2/"+filepath.Base(configFileYaml.Name()), os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		configFileYaml.Close()
		configFileOverloaded1.Close()
		configFileOverloaded2.Close()
		os.RemoveAll(configFileYaml.Name())
		os.RemoveAll(configFileOverloaded1.Name())
		os.RemoveAll(configFileOverloaded2.Name())
	}()

	// Write Config to File
	configFileYaml.WriteString(`
nested:
    key:
        a: AA
        b: BB
    `)

	configFileOverloaded1.WriteString(`
nested:
    key:
        a: XX
        b: YY
    `)

	configFileOverloaded2.WriteString(`
nested:
    key:
        a: MM
        b: NN
    `)

	_ = os.Setenv("CONFIG_DIR", configFileOverloaded1.Name())
	_ = os.Setenv("APPNAME_CONFIG_DIR", configFileOverloaded2.Name())

	configLoaderEnvDefault, err := configuro.NewConfig(
		configuro.WithoutLoadFromEnvVars(),
		configuro.WithoutLoadDotEnv(),
		configuro.WithLoadFromConfigFile(configFileYaml.Name(), false),
	)
	if err != nil {
		t.Fatal(err)
	}

	configLoaderWithEnvDir, err := configuro.NewConfig(
		configuro.WithoutLoadFromEnvVars(),
		configuro.WithoutLoadDotEnv(),
		configuro.WithLoadFromConfigFile(configFileYaml.Name(), false),
		configuro.WithEnvConfigPathOverload("APPNAME_CONFIG_DIR"),
	)
	if err != nil {
		t.Fatal(err)
	}

	tests := []test{
		{name: "configLoaderEnvDirDefault", config: configLoaderEnvDefault, expected: Example{
			Nested: Nested{
				Key: Key{
					A: "XX",
					B: "YY",
				},
			},
		}},
		{name: "configLoaderWithEnvDir", config: configLoaderWithEnvDir, expected: Example{
			Nested: Nested{
				Key: Key{
					A: "MM",
					B: "NN",
				},
			},
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			example := &Example{}
			err := test.config.Load(example)
			if err != nil {
				t.Fatal(err)
			}

			if example.Nested.Key.A != test.expected.Nested.Key.A ||
				example.Nested.Key.B != test.expected.Nested.Key.B {
				t.Fatalf("Loaded Values doesn't equal expected values. loaded: %v, expected: %v", example, test.expected)
			}
		})
	}
}

func TestExpandEnvVar(t *testing.T) {
	configFileYaml, err := ioutil.TempFile("", "TestExpandEnvVar*.yml")
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		configFileYaml.Close()
		os.RemoveAll(configFileYaml.Name())
	}()

	_ = os.Setenv("KEY_A", "AA")
	_ = os.Setenv("KEY_B", "B")
	_ = os.Setenv("KEY_D", "")
	_ = os.Setenv("OBJECT", `{"a":123, "b": "abc"}`)
	_ = os.Setenv("NUMBER", "123456")
	_ = os.Setenv("NUMBERLIST1", "1,2,3")
	_ = os.Setenv("NUMBERLIST2", "[\"4\",5,6]")
	_ = os.Setenv("INTMAP", `{"a":123, "b": "456"}`)

	// Write Config to File
	configFileYaml.WriteString(`
nested:
    key:
        a: ${KEY_A}
        b: A${KEY_B}C
        c: ${KEY_C|defaultC}
        d: ${KEY_D|defaultD}
        e: ${KEY_E|}
    key_a: ${OBJECT}
    number: ${NUMBER}
    numberList1: ${NUMBERLIST1}
    numberList2: ${NUMBERLIST2}
    IntMap: ${INTMAP}
    `)

	envOnlyWithoutPrefix, err := configuro.NewConfig(
		configuro.WithExpandEnvVars(),
		configuro.WithLoadFromEnvVars("X"),
		configuro.WithLoadDotEnv(""),
		configuro.WithLoadFromConfigFile(configFileYaml.Name(), false),
		configuro.WithEnvConfigPathOverload(""),
	)
	if err != nil {
		t.Fatal(err)
	}

	tests := []test{
		{name: "withoutPrefix", config: envOnlyWithoutPrefix, expected: Example{
			Nested: Nested{
				Key: Key{
					A:     "AA",
					B:     "ABC",
					C:     "defaultC",
					D:     "",
					E:     "",
					EMPTY: "",
				},
				Key_A: &Key{
					A: "123",
					B: "abc",
				},
				Number:      123456,
				NumberList1: []int{1, 2, 3},
				NumberList2: []int{4, 5, 6},
				IntMap:      map[string]int{"a": 123, "b": 456},
			},
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			example := &Example{}
			err := test.config.Load(example)
			if err != nil {
				t.Fatal(err)
			}

			if example.Nested.Key.A != test.expected.Nested.Key.A ||
				example.Nested.Key.B != test.expected.Nested.Key.B ||
				example.Nested.Key.C != test.expected.Nested.Key.C ||
				example.Nested.Key.D != test.expected.Nested.Key.D ||
				example.Nested.Key.E != test.expected.Nested.Key.E ||
				example.Nested.Key.EMPTY != test.expected.Nested.Key.EMPTY ||
				example.Nested.Key_A.A != test.expected.Nested.Key_A.A ||
				example.Nested.Key_A.B != test.expected.Nested.Key_A.B ||
				example.Nested.Number != test.expected.Nested.Number ||
				example.Nested.IntMap["a"] != test.expected.Nested.IntMap["a"] ||
				example.Nested.IntMap["b"] != test.expected.Nested.IntMap["b"] ||
				!equalSlice(example.Nested.NumberList1, test.expected.Nested.NumberList1) ||
				!equalSlice(example.Nested.NumberList2, test.expected.Nested.NumberList2) {
				t.Fatalf("Loaded Values doesn't equal expected values. loaded: %v, expected: %v", example, test.expected)
			}
		})
	}
}

func TestChangeTagName(t *testing.T) {
	configFileYaml, err := ioutil.TempFile("", "TestChangeTagName*.yml")
	if err != nil {
		t.Fatal(err)
	}
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

	configLoaderDefaultTag, err := configuro.NewConfig(
		configuro.WithLoadFromEnvVars("X"),
		configuro.WithLoadDotEnv(""),
		configuro.WithLoadFromConfigFile(configFileYaml.Name(), false),
		configuro.WithEnvConfigPathOverload(""),
	)
	if err != nil {
		t.Fatal(err)
	}

	configLoaderNewTag, err := configuro.NewConfig(
		configuro.Tag("newtag", "validate"),
		configuro.WithLoadFromEnvVars("X"),
		configuro.WithLoadDotEnv(""),
		configuro.WithLoadFromConfigFile(configFileYaml.Name(), false),
		configuro.WithEnvConfigPathOverload(""),
	)
	if err != nil {
		t.Fatal(err)
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
				t.Fatal(err)
			}

			if example.Object.KeyA != test.expected.Object.KeyA ||
				example.Object.KeyB != test.expected.Object.KeyB {
				t.Fatalf("Loaded Values doesn't equal expected values. loaded: %v, expected: %v", example, test.expected)
			}
		})
	}
}

func TestValidateByTag(t *testing.T) {
	configFileYaml, err := ioutil.TempFile("", "TestValidateByTag*.yml")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		configFileYaml.Close()
		os.RemoveAll(configFileYaml.Name())
	}()

	// Write Config to File
	configFileYaml.WriteString(`
nested:
    key:
        a: A
        b: A
        c: X
    `)

	configLoader, err := configuro.NewConfig(
		configuro.WithValidateByTags(),
		configuro.WithLoadFromEnvVars("X"),
		configuro.WithLoadDotEnv(""),
		configuro.WithLoadFromConfigFile(configFileYaml.Name(), false),
		configuro.WithEnvConfigPathOverload(""),
	)
	if err != nil {
		t.Fatal(err)
	}

	example := &Example{}
	err = configLoader.Load(example)
	if err != nil {
		t.Fatal(err)
	}

	err = configLoader.Validate(example)
	if err == nil {
		t.Fatal("Validation with Tags was bypassed.")
	} else {
		_ = err.Error()
		errx, ok := err.(*configuro.ErrValidationTag)
		if !ok {
			t.Fatal("Validation with Tags Returned Wrong Error Type.")
		}
		_ = errx.Unwrap()
	}
}

func TestValidateByTagMultiErr(t *testing.T) {
	configFileYaml, err := ioutil.TempFile("", "TestValidateByTag*.yml")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		configFileYaml.Close()
		os.RemoveAll(configFileYaml.Name())
	}()

	// Write Config to File
	configFileYaml.WriteString(`
nested:
    key:
        a: A
        b: A
    `)

	configLoader, err := configuro.NewConfig(
		configuro.WithValidateByTags(),
		configuro.WithoutValidateByFunc(),
		configuro.WithLoadFromEnvVars("X"),
		configuro.WithLoadDotEnv(""),
		configuro.WithLoadFromConfigFile(configFileYaml.Name(), false),
		configuro.WithEnvConfigPathOverload(""),
	)
	if err != nil {
		t.Fatal(err)
	}

	example := &Example{}
	err = configLoader.Load(example)
	if err != nil {
		t.Fatal(err)
	}

	err = configLoader.Validate(example)
	if err == nil {
		t.Fatal("Validation with Tags was bypassed.")
	}
	errsType, ok := err.(configuro.ErrValidationErrors)
	if !ok {
		t.Fatal("Validation with Tags didn't return ErrValidationErrors when multi error happen.")

	}

	Errs := errsType.Errors()

	for _, err := range Errs {
		_, ok := err.(*configuro.ErrValidationTag)
		if !ok {
			t.Fatal("Validation with Tags Returned Wrong Error Type.")
		}
	}
	_ = errsType.Unwrap()
	_ = err.Error()
}

func TestValidateByInterface(t *testing.T) {
	configFileYaml, err := ioutil.TempFile("", "TestValidateByInterface*.yml")
	if err != nil {
		t.Fatal(err)
	}

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
    `)

	configLoader, err := configuro.NewConfig(
		configuro.WithoutValidateByTags(),
		configuro.WithValidateByFunc(true, true),
		configuro.WithLoadFromEnvVars("X"),
		configuro.WithLoadDotEnv(""),
		configuro.WithLoadFromConfigFile(configFileYaml.Name(), false),
		configuro.WithEnvConfigPathOverload(""),
	)
	if err != nil {
		t.Fatal(err)
	}

	example := &Example{}
	err = configLoader.Load(example)
	if err != nil {
		t.Fatal(err)
	}

	err = configLoader.Validate(example)
	Errs := multierr.Errors(err)
	if Errs == nil {
		t.Fatal("Validation using validator interface was bypassed.")
	} else {
		for _, err := range Errs {
			errx, ok := err.(*configuro.ErrValidationFunc)
			if !ok {
				t.Fatal("Validation using validator interface Returned Wrong Error Type.")
			}
			_ = errx.Unwrap()
		}
	}

	_ = err.Error()
}

func TestValidateMaps(t *testing.T) {
	configFileYaml, err := ioutil.TempFile("", "TestValidateMaps*.yml")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		configFileYaml.Close()
		os.RemoveAll(configFileYaml.Name())
	}()

	// Write Config to File
	configFileYaml.WriteString(`
nested:
  KeyMap:
    One:
      a: ONE
      b: ONE
    Two:
      a: ONE
      b: TWO
    `)

	configLoader, err := configuro.NewConfig(
		configuro.WithoutValidateByTags(),
		configuro.WithLoadFromEnvVars("X"),
		configuro.WithLoadDotEnv(""),
		configuro.WithLoadFromConfigFile(configFileYaml.Name(), false),
		configuro.WithEnvConfigPathOverload(""),
	)
	if err != nil {
		t.Fatal(err)
	}

	example := &Example{}
	err = configLoader.Load(example)
	if err != nil {
		t.Fatal(err)
	}

	err = configLoader.Validate(example)
	Errs := multierr.Errors(err)
	if Errs == nil {
		t.Fatal("Validation using validator interface was bypassed.")
	} else {
		for _, err := range Errs {
			_, ok := err.(*configuro.ErrValidationFunc)
			if !ok {
				t.Fatal("Validation using validator interface Returned Wrong Error Type.")
			}
		}
	}
}

func TestValidateSlices(t *testing.T) {
	configFileYaml, err := ioutil.TempFile("", "TestValidateSlices*.yml")
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		configFileYaml.Close()
		os.RemoveAll(configFileYaml.Name())
	}()

	// Write Config to File
	configFileYaml.WriteString(`
nested:
  KeySlice:
    - a: ONE
      b: ONE
    - a: ONE
      b: TWO
    `)

	configLoader, err := configuro.NewConfig(
		configuro.WithoutValidateByTags(),
		configuro.WithLoadFromEnvVars("X"),
		configuro.WithLoadDotEnv(""),
		configuro.WithLoadFromConfigFile(configFileYaml.Name(), false),
		configuro.WithEnvConfigPathOverload(""),
	)
	if err != nil {
		t.Fatal(err)
	}

	example := &Example{}
	err = configLoader.Load(example)
	if err != nil {
		t.Fatal(err)
	}

	err = configLoader.Validate(example)
	Errs := multierr.Errors(err)
	if Errs == nil {
		t.Fatal("Validation using validator interface was bypassed.")
	} else {
		for _, err := range Errs {
			_, ok := err.(*configuro.ErrValidationFunc)
			if !ok {
				t.Fatal("Validation using validator interface Returned Wrong Error Type.")
			}
		}
	}
}

func equalSlice(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
