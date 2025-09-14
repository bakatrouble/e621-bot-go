package e621

type PostFile struct {
	Width  int     `json:"width"`
	Height int     `json:"height"`
	Ext    string  `json:"ext"`
	Size   int     `json:"size"`
	Md5    string  `json:"md5"`
	Url    *string `json:"url,omitempty"`
}

type PostTags struct {
	General   []string `json:"general"`
	Species   []string `json:"species"`
	Character []string `json:"character"`
	Copyright []string `json:"copyright"`
	Artist    []string `json:"artist"`
	Invalid   []string `json:"invalid"`
	Lore      []string `json:"lore"`
	Meta      []string `json:"meta"`
}

type Post struct {
	ID        int      `json:"id"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
	File      PostFile `json:"file"`
	Tags      PostTags `json:"tags"`
}

func (p *Post) FlatTags() []string {
	return append(
		append(
			append(
				append(
					append(
						append(
							append(
								p.Tags.General, p.Tags.Species...,
							),
							p.Tags.Character...,
						),
						p.Tags.Copyright...,
					),
					p.Tags.Artist...,
				),
				p.Tags.Invalid...,
			),
			p.Tags.Lore...,
		),
		p.Tags.Meta...,
	)
}

func (p *Post) FlatTagsMap() map[string]bool {
	tags := make(map[string]bool)
	for _, tag := range p.FlatTags() {
		tags[tag] = true
	}
	return tags
}

type PostVersion struct {
	ID          int      `json:"id"`
	PostID      int      `json:"post_id"`
	Tags        string   `json:"tags"`
	AddedTags   []string `json:"added_tags"`
	RemovedTags []string `json:"removed_tags"`
}

type TagAlias struct {
	Status         string `json:"status"`
	AntecedentName string `json:"antecedent_name"`
	ConsequentName string `json:"consequent_name"`
}
