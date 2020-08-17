%{
package parser

import (
  "fmt"
  "bytes"
  "io"
  "bufio"
	"os"
	"MIA-P1/actions"
)

type node struct {
  name string
  children []node
}

func (n node) String() string {
  buf := new(bytes.Buffer)
  n.print(buf, " ")
  return buf.String()
}

func (n node) print(out io.Writer, indent string) {
  fmt.Fprintf(out, "\n%v%v", indent, n.name)
  for _, nn := range n.children { nn.print(out, indent + "  ") }
}

func Node(name string) node { return node{name: name} }
func (n node) append(nn...node) node { n.children = append(n.children, nn...); return n }


%}

%union{
    node node
    token string
}

// Terminals 

%token <token> arrow
%token <token> add
%token <token> delete
%token <token> digit
%token <token> digits
%token <token> diskName
%token <token> equals
%token <token> exec
%token <token> fit
%token <token> fdisk
%token <token> greater
%token <token> hyphen
%token <token> idn 
%token <token> id
%token <token> less
%token <token> mia_file
%token <token> mount
%token <token> mount_name
%token <token> mbr
%token <token> mkfs
%token <token> mkdisk
%token <token> number
%token <token> name
%token <token> path
%token <token> pause
%token <token> rmdisk
%token <token> size
%token <token> tpe
%token <token> unit
%token <token> unmount
%token <token> route
%token <token> quote


%type <token> add
%type <token> arrow
%type <token> delete
%type <token> digit
%type <token> digits
%type <token> diskName
%type <token> equals
%type <token> exec
%type <token> fit
%type <token> fdisk
%type <token> greater
%type <token> hyphen
%type <token> idn 
%type <token> id
%type <token> less
%type <token> mia_file
%type <token> mount
%type <token> mount_name
%type <token> mbr
%type <token> mkfs
%type <token> mkdisk
%type <token> number
%type <token> name
%type <token> path
%type <token> pause
%type <token> rmdisk
%type <token> size
%type <token> tpe
%type <token> unit
%type <token> unmount
%type <token> route
%type <token> quote



// Non terminals
%type <node> MOUNT
%type <node> PAUSE
%type <node> UNMOUNT
%type <node> EXEC
%type <node> MKDISK
%type <node> RMDISK
%type <node> FDISK
%type <node> MKFS
%type <node> INSTRUCTION
%type <node> INSTRUCTIONS



%%

start: INSTRUCTIONS;

INSTRUCTIONS: INSTRUCTION
			| INSTRUCTIONS INSTRUCTION;

INSTRUCTION: MOUNT
			| UNMOUNT
			| EXEC
			| MKDISK
			| RMDISK
			| FDISK
			| MKFS
			| PAUSE;

PAUSE: pause { actions.PauseAction() };

EXEC: exec hyphen path arrow route {actions.GetFile($5)};

MKDISK: mkdisk hyphen size arrow digit hyphen path arrow quote route quote hyphen name arrow diskName {actions.MkdiskCreateRoute($5, $10, $15)}
;

MOUNT: mount hyphen path arrow route hyphen name arrow mount_name {$$ = Node($1)};

UNMOUNT: unmount hyphen idn {$$ = Node($1)};

FDISK: fdisk hyphen size arrow digit {$$ = Node($1)}
;

RMDISK: rmdisk hyphen path arrow route {$$ = Node($1)};



MKFS: mkfs hyphen id arrow idn {$$ = Node($1)}
	| mkfs hyphen id arrow idn hyphen tpe arrow id {$$ = Node($1)}
;


%% 

// Run exported
func Run() {
	fi := bufio.NewReader(os.NewFile(0, "stdin"))
	yyDebug = 0
	yyErrorVerbose = true
	for {
		var eqn string
		var ok bool

		fmt.Printf(">> ")
		if eqn, ok = input(fi); ok {
			l := newLexer(bytes.NewBufferString(eqn), os.Stdout, "file.name")
			yyParse(l)
		} else {
			break
		}
	}

}

func input(fi *bufio.Reader) (string, bool) {
	s, err := fi.ReadString('\n')
	if err != nil {
		return "", false
	}
	return s, true
}
