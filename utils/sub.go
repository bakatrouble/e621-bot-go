package utils

type Sub struct {
	Raw      string
	tag      string
	negative bool
	children []Sub
}

//func parseSub(sub string) Sub {
//	result := Sub{Raw: sub}
//
//	parts := strings.Split(sub, " ")
//	if len(parts) > 1 {
//		for _, part := range parts {
//			result.children = append(result.children, parseSub(part))
//		}
//		return result
//	}
//
//	result.tag = sub
//	if sub[0] == '-' {
//		result.tag = sub[1:]
//		result.negative = true
//	}
//
//	return Sub{
//		Raw:      sub,
//		tag:      &sub,
//		negative: negative,
//	}
//}
//
//func ParseSubs(subs []string) (result []Sub) {
//	for _, rawSub := range subs {
//		if rawSub == "" {
//			continue
//		}
//		sub := rawSub
//		negative := false
//		if rawSub[0] == '-' {
//			negative = true
//			sub = rawSub[1:]
//		}
//		result = append(result, Sub{
//			Raw:      sub,
//			tag:      tag,
//			negative: negative,
//		})
//	}
//	return
//}
