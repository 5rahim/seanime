package seanime_parser

type (
	enclosedGroup struct {
		openingBracket *token
		closingBracket *token
		tokens         []*token
	}
)

func (tm *tokenManager) getEnclosedGroups() []*enclosedGroup {
	//t := tm.tokens

	//enclosedGroups := make([]*enclosedGroup, 0)
	//for _, tkn := range *t {
	//	if !tkn.isEnclosed() {
	//		continue
	//	}
	//
	//}

	return nil
}
