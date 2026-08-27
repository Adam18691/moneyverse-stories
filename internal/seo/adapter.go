package seo

// Metadata: توافق مع main.go — نفس حقول Meta بدون json tags
type Metadata struct {
	Title       string
	Description string
	Tags        []string
}

// GenerateMetadata: توافق مع main.go — يلف Generate
func GenerateMetadata(story string) (*Metadata, error) {
	m := Generate(story, story, story)
	return &Metadata{
		Title:       m.Title,
		Description: m.Description,
		Tags:        m.Tags,
	}, nil
}
