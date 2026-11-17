package modinfo

import "testing"

func FuzzParse(f *testing.F) {
	f.Add([]byte(`{
		// The game reads this with Newtonsoft, which takes comments and trailing commas.
		"Type": "code", "ModID": "carryon", "Name": "Carry On", "Version": "1.14.3",
		"authors": ["copygirl", "Nikea"],
		"Dependencies": { "game": "1.22.0", "VSImGui": "" },
	}`))
	f.Add([]byte(`{"name":"Foo","modid":"foo","version":"1.0.0"}`))
	f.Add([]byte(`{"name":"Foo","modid":"foo","version":"1.0.0","authors":["a"],"dependencies":{"game":"1.21.0","bar":"1.0"}}`))
	f.Add([]byte(`{/* comment */ "Name": "X",}`))
	f.Add([]byte(`{"modid": "a", "version": 1.2}`))
	f.Add([]byte(`{"dependencies":{"game":1.21}}`))
	f.Add([]byte(`{"dependencies":{"game":" * "}}`))
	f.Add([]byte(`{"dependencies":{"game":"*"}}`))
	f.Add([]byte(`{"dependencies":{"game":"v1.20.0"}}`))
	f.Add([]byte(`{"modid":"lower","ModID":"UPPER"}`))
	f.Add([]byte(`{"dependencies":{"game":"1.20","Game":"1.21"}}`))
	f.Add([]byte(`[]`))
	f.Add([]byte(`null`))
	f.Add([]byte(``))
	f.Fuzz(func(t *testing.T, b []byte) {
		info, err := Parse(b)
		if err != nil {
			return
		}
		n := len(b)
		if len(info.Dependencies) > n || len(info.Authors) > n {
			t.Fatalf("slice longer than input: deps=%d authors=%d n=%d", len(info.Dependencies), len(info.Authors), n)
		}
		if len(info.Name) > n || len(info.ModID) > n || len(info.Version) > n || len(info.GameVersion) > n {
			t.Fatalf("field longer than input")
		}
	})
}
