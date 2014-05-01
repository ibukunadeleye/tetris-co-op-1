package storageServer


type PreparePkg struct{
	N int
	CS int
	HostPort string
}

type AcceptPkg struct{
	N int
	V []byte
	CS int	
	HostPort string
}

type CommitPkg struct{
	N int
	V []byte
	hostport string
}

type Reply struct{
	ok bool
	value []byte
	CS int
}