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

var newDisk actions.Disk = actions.Disk{}
var newPartition actions.Partition = actions.Partition{}
var newFDisk actions.FDISK = actions.FDISK{}

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
%token <token> read
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
%type <token> read
%type <token> route
%type <token> quote



// Non terminals
%type <node> MOUNT
%type <node> MOUNTO
%type <node> MOUNTT
%type <node> PAUSE
%type <node> UNMOUNT
%type <node> EXEC
%type <node> MKDISK
%type <node> MKDISKO
%type <node> MKDISKT
%type <node> RMDISK
%type <node> FDISK
%type <node> FDISKO
%type <node> FDISKT
%type <node> MKFS
%type <node> INSTRUCTION
%type <node> INSTRUCTIONS
%type <node> READ



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
			| PAUSE
			| READ
			;

PAUSE: pause { actions.PauseAction() };

READ: read arrow route { actions.ReadFile($3)}

EXEC: exec hyphen path arrow route {actions.GetFile($5)};

MKDISK: mkdisk MKDISKO { 
	newDisk.CreateDisk()
	newDisk = actions.Disk{}
 }

MKDISKO: MKDISKT 
| MKDISKO MKDISKT EMPTY;

MKDISKT: hyphen size arrow digit { newDisk.SetDiskSize($4) }
| hyphen path arrow quote route quote { newDisk.SetDiskRoute($5) }
| hyphen name arrow diskName { newDisk.SetDiskName($4) }
| hyphen unit arrow id { newDisk.SetDiskUnit($4) }
;



MOUNT: mount MOUNTO { $$ = Node($1) };

MOUNTO: MOUNTT
	|	MOUNTO MOUNTT EMPTY;

MOUNTT: hyphen path arrow route { actions.PrintParameter($4) }
	|	hyphen name arrow id { actions.PrintParameter($4) };

UNMOUNT: unmount hyphen idn {$$ = Node($1)};

FDISK: fdisk FDISKO {
	newFDisk.CreatePartition()
	newPartition = actions.Partition{}
	newFDisk = actions.FDISK{}
}
;

FDISKO: FDISKT 
| FDISKO FDISKT EMPTY;

FDISKT:	hyphen unit arrow id  { newFDisk.SetFUnit($4) }
| hyphen tpe arrow id { newFDisk.SetPartitionType($4) }
| hyphen fit arrow id { newFDisk.SetPartitionFit($4) }
| hyphen delete arrow id { newFDisk.SetDeleteOption($4) }
| hyphen add { actions.PrintParameter($2) }
| hyphen size arrow digit  { newFDisk.SetPSize($4) }
| hyphen name arrow id { newFDisk.SetPartitionName($4) }
| hyphen path arrow quote route quote { newFDisk.SetPartitionRoute($5) }
;


RMDISK: rmdisk hyphen path arrow quote route quote {actions.RemoveDisk($6)};



MKFS: mkfs hyphen id arrow idn {$$ = Node($1)}
	| mkfs hyphen id arrow idn hyphen tpe arrow id {$$ = Node($1)}
;

EMPTY: ;


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
