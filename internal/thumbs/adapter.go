package thumbs

// Generate: توافق مع main.go — يلف SmartGenerate
func Generate(id int, text, fallbackImg string) error {
	_, err := SmartGenerate(id, text, text, text)
	return err
}
