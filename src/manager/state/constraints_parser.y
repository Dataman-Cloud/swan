%{
package state

import (
	"strings"
        "errors"
	"text/scanner"
)

var keywords = map[string]int{
  "and": AND,
  "or": OR,
  "unique": UNIQUE,
  "like": LIKE,
  "contains": CONTAINS,
  "not": NOT,
  "equal": EQUAL,
}
%}

%union{
    token string
    what  string
    param string
    str string
    expr  Statement
}

%start program

%type <expr> program
%type <expr> expr

%token <token> AND OR UNIQUE LIKE CONTAINS NOT EQUAL
%token <str> IDENTIFIER

%type <what> what
%type <param> param

%%

program
    : expr
    {
        $$ = $1
        yylex.(*ConstraintParser).result = $$
    }

expr
    : AND '(' expr ')' '(' expr ')' { $$ = &AndStatement{Op1: $3, Op2: $6} }
    | OR '(' expr ')' '(' expr ')' { $$ = &OrStatement{Op1: $3, Op2: $6} }
    | NOT '(' expr ')' { $$ = &NotStatement{Op1: $3} }
    | UNIQUE what { $$ = &UniqueStatment{What: $2}; }
    | LIKE what param { $$ = &LikeStatement{What: $2, Regex: $3} }
    | EQUAL what param { $$ = &EqualStatement{What: $2, Regex: $3} }
    | CONTAINS what param { $$ = &LikeStatement{What: $2, Regex: $3} }
    ;
what: IDENTIFIER { $$ = $1; }
param: IDENTIFIER { $$ = $1 }


%%

type ConstraintParser struct {
	scanner.Scanner
	result Statement
	hasErrors bool
}

func (l *ConstraintParser) Lex(lval *yySymType) int {
	tok := l.Scan()
	switch tok {
	case scanner.Ident:
		ident := l.TokenText()
		keyword, isKeyword := keywords[ident]
		if isKeyword {
			return keyword
		}
		lval.str = ident
		return IDENTIFIER
	case scanner.String:
		text := l.TokenText()
		text = text[1 : len(text)-1]
		lval.str = text
		return IDENTIFIER
	default:
		return int(tok)
	}
}

func (l *ConstraintParser) Error(e string) {
  l.hasErrors = true
}

func ParseConstraint(c string) (Statement, error) {
	l := new(ConstraintParser)
	l.Init(strings.NewReader(c))
	yyParse(l)
  if l.hasErrors {
    return nil, errors.New("parse error")
  }else{
    return l.result, nil
  }
}


