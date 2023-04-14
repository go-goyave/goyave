package config

// TMP file so the rest of the project still compiles
// TODO remove this

func IsLoaded() bool {
	return false
}

func Load() error {
	return nil
}

func LoadFrom(_ string) error {
	return nil
}

func Get(_ string) any {
	return ""
}

func GetString(_ string) string {
	return ""
}

func GetBool(_ string) bool {
	return false
}

func GetFloat(_ string) float64 {
	return 0
}

func GetInt(_ string) int {
	return 0
}

func Has(_ string) bool {
	return false
}

func Set(_ string, _ any) {}

func Clear() {}
