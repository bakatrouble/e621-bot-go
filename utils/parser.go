package utils

import (
	"context"
	"e621-bot-go/storage"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

type Atom struct {
	//Rating *string `  ("rating" ":" @("s" | "q" | "e"))`
	Tag *string `@Ident`
}

func (a *Atom) Check(tagList map[string]bool) bool {
	if a.Tag != nil {
		value, exists := tagList[*a.Tag]
		return exists && value == true
	}
	return true
}

type Term struct {
	Not      *Term  `  "-" @@`
	Atom     *Atom  `| @@`
	SubQuery *Query `| "{" @@ "}"`
}

func (t *Term) Check(tagList map[string]bool) bool {
	if t.Not != nil {
		return !t.Not.Check(tagList)
	} else if t.Atom != nil {
		return t.Atom.Check(tagList)
	} else if t.SubQuery != nil {
		return t.SubQuery.Check(tagList)
	}
	return true
}

func (t *Term) MentionedTags() map[string]bool {
	if t.Not != nil {
		return t.Not.MentionedTags()
	} else if t.Atom != nil {
		return map[string]bool{*t.Atom.Tag: true}
	} else if t.SubQuery != nil {
		return t.SubQuery.MentionedTags()
	}
	return map[string]bool{}
}

type Query struct {
	And  []*Term `  @@ @@+`
	Or   []*Term `| @@ ("|" @@)+`
	Term *Term   `| @@`
}

func (q *Query) Check(tagList map[string]bool) bool {
	if q.And != nil {
		for _, term := range q.And {
			if !term.Check(tagList) {
				return false
			}
		}
		return true
	} else if q.Or != nil {
		for _, term := range q.Or {
			if term.Check(tagList) {
				return true
			}
		}
		return false
	} else if q.Term != nil {
		return q.Term.Check(tagList)
	}
	return true
}

func (q *Query) MentionedTags() map[string]bool {
	tagSet := make(map[string]bool)
	if q.And != nil {
		for _, term := range q.And {
			for tag, _ := range term.MentionedTags() {
				tagSet[tag] = true
			}
		}
	} else if q.Or != nil {
		for _, term := range q.Or {
			for tag, _ := range term.MentionedTags() {
				tagSet[tag] = true
			}
		}
	} else if q.Term != nil {
		for tag, _ := range q.Term.MentionedTags() {
			tagSet[tag] = true
		}
	}
	return tagSet
}

var l = lexer.MustSimple([]lexer.SimpleRule{
	{"Ident", `[a-zA-Z0-9_][^\s{}|]*`},
	{"Punct", `[-{}|:]`},
	{"whitespace", `\s+`},
})
var QueryParser = participle.MustBuild[Query](
	participle.UseLookahead(1024),
	participle.Lexer(l),
)

type QueryInfo struct {
	Raw   string
	Query *Query
}

var EmptyQueryInfo = &QueryInfo{"", &Query{}}

func (qi *QueryInfo) Check(tagList map[string]bool) bool {
	if qi.Query == nil {
		return true
	}
	return qi.Query.Check(tagList)
}

func ParseSubs(subs []string) (result []*QueryInfo, err error) {
	for _, sub := range subs {
		if sub == "" {
			continue
		}
		query, err := QueryParser.ParseString(sub, sub)
		if err != nil {
			return nil, err
		}
		result = append(result, &QueryInfo{sub, query})
	}
	return result, nil
}

func GetQueries(store *storage.Storage, ctx context.Context) ([]*QueryInfo, error) {
	subs, err := store.GetSubs(ctx)
	if err != nil {
		return nil, err
	}

	queries, err := ParseSubs(subs)
	if err != nil {
		return nil, err
	}
	return queries, nil
}
