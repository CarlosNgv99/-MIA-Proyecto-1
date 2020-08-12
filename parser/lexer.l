%{

package parser

import (
  "bufio"
  "go/token"
  "io"
	"unicode"

	"modernc.org/golex/lex"
)

// Allocate Character classes anywhere in [0x80, 0xFF].
const (
	classUnicodeLeter = iota + 0x80
	classUnicodeDigit
	classOther
)

type lexer struct {
	*lex.Lexer
}

func rune2Class(r rune) int {
	if r >= 0 && r < 0x80 { // Keep ASCII as it is.
		return int(r)
	}
	if unicode.IsLetter(r) {
		return classUnicodeLeter
	}
	if unicode.IsDigit(r) {
		return classUnicodeDigit
	}
	return classOther
}

func newLexer(src io.Reader, dst io.Writer, fName string) *lexer {
  file := token.NewFileSet().AddFile(fName, -1, 1<<31-1)
  lx, err := lex.New(file, bufio.NewReader(src), lex.RuneClass(rune2Class))
  if (err != nil) { panic(err) }
  return &lexer{lx}
}

func (l *lexer) Lex(lval *yySymType) int {
  c := l.Enter()

%}
%option case-insensitive
%yyc c
%yyn c = l.Next()
%yym l.Mark()

digit         [0-9]|{unicodeDigit}
digits		  [0-9]*|{unicodeDigit}*
id    {letter}({letter}|{digit})*
letter        [_a-zA-Z]|{unicodeLetter}
diskName	  {letter}({letter}|{digit})*"."disk
route         "/"("/"|"-"|"_"|"."|[a-zA-Z0-9])+
arrow           "-"">"
hyphen          "-"
quote			\"
unicodeDigit  \x81 
unicodeLetter \x80
%%
	c = l.Rule0()

[ \t\r\n]+
pause  lval.token = string(l.TokenBytes(nil)); return pause  
exec  lval.token = string(l.TokenBytes(nil)); return exec  
path  lval.token = string(l.TokenBytes(nil)); return path  
mount  lval.token = string(l.TokenBytes(nil)); return mount  
unmount  lval.token = string(l.TokenBytes(nil)); return unmount  
hyphen  lval.token = string(l.TokenBytes(nil)); return hyphen  
arrow  lval.token = string(l.TokenBytes(nil)); return arrow  
name  lval.token = string(l.TokenBytes(nil)); return name  
size  lval.token = string(l.TokenBytes(nil)); return size  
mkdisk  lval.token = string(l.TokenBytes(nil)); return mkdisk  
{id} lval.token = string(l.TokenBytes(nil)); return id 
{route} lval.token = string(l.TokenBytes(nil)); return route
{hyphen} lval.token = string(l.TokenBytes(nil)); return hyphen
{arrow} lval.token = string(l.TokenBytes(nil)); return arrow
{digit} lval.token = string(l.TokenBytes(nil)); return digit
{digits} lval.token = string(l.TokenBytes(nil)); return digits
{quote} lval.token = string(l.TokenBytes(nil)); return quote
{diskName} lval.token = string(l.TokenBytes(nil)); return diskName

%%
	if c, ok := l.Abort(); ok { return int(c) }
	goto yyAction
}