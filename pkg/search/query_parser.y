%{
package search

import (
	"strings"

	"github.com/blevesearch/bleve/search/query"
)

%}

%union{
  term string
  termList []string
  boolean *query.BooleanQuery
  query query.Query
}

%token tokError tokEOF tokTerm
%type <term> tokTerm 
%type <termList> Words
%type <query> Query Match
%type <boolean> main Boolean BooleanPart

%start main

%%

main: Boolean
	{ setResult(yylex, $1) }

Boolean
	: BooleanPart
	| Boolean BooleanPart
	{ $$ = mergeBoolean($1, $2) }

BooleanPart
	: Query
	{ $$ = query.NewBooleanQueryForQueryString(nil, []query.Query{$1}, nil) }
	| '+' Query
	{ $$ = query.NewBooleanQueryForQueryString([]query.Query{$2}, nil, nil) }
	| '-' Query
	{ $$ = query.NewBooleanQueryForQueryString(nil, nil, []query.Query{$2}) }

Query
	: Match
	{ $$ = $1 }
	| tokTerm ':' Match
	{ $$ = setField($3, $1) }
	| '#' tokTerm
	{ $$ = setField(query.NewMatchQuery($2), "tags") }

Match
	: tokTerm
	{ $$ = query.NewMatchQuery($1) }
	| '"' Words '"'
	{ $$ = query.NewMatchPhraseQuery(strings.Join($2, " ")) }
	| '`' Words '`'
	{ $$ = query.NewPhraseQuery($2, "_all") }

Words
	: tokTerm
	{ $$ = []string{$1} }
	| Words tokTerm
	{ $$ = append($1, $2) }

// vim:sw=4:noexpandtab:nolist
