module cope-file

require (
	github.com/lxn/walk v0.0.0-20191001144247-31870cf268b0
	github.com/lxn/win v0.0.0-20190919090605-24c5960b03d8 // indirect
	gopkg.in/Knetic/govaluate.v3 v3.0.0 // indirect
)

replace (
	cloud.google.com/go => github.com/GoogleCloudPlatform/google-cloud-go v0.44.3
	cloud.google.com/go/datastore => github.com/GoogleCloudPlatform/google-cloud-go/datastore v1.0.0
	golang.org/x/crypto => github.com/golang/crypto v0.0.0-20190829043050-9756ffdc2472
	golang.org/x/exp => github.com/golang/exp v0.0.0-20190829153037-c13cbed26979
	golang.org/x/image => github.com/golang/image v0.0.0-20190902063713-cb417be4ba39
	golang.org/x/lint => github.com/golang/lint v0.0.0-20190409202823-959b441ac422
	golang.org/x/mobile => github.com/golang/mobile v0.0.0-20190830201351-c6da95954960
	golang.org/x/net => github.com/golang/net v0.0.0-20190827160401-ba9fcec4b297
	golang.org/x/oauth2 => github.com/golang/oauth2 v0.0.0-20190604053449-0f29369cfe45
	golang.org/x/sync => github.com/golang/sync v0.0.0-20190423024810-112230192c58
	golang.org/x/sys => github.com/golang/sys v0.0.0-20190902133755-9109b7679e13
	golang.org/x/text => github.com/golang/text v0.3.2
	golang.org/x/time => github.com/golang/time v0.0.0-20190308202827-9d24e82272b4
	golang.org/x/tools => github.com/golang/tools v0.0.0-20190903025054-afe7f8212f0d
	golang.org/x/xerrors => github.com/golang/xerrors v0.0.0-20190717185122-a985d3407aa7
	google.golang.org/api => github.com/googleapis/google-api-go-client v0.9.0
	google.golang.org/appengine => github.com/golang/appengine v1.6.2
	google.golang.org/genproto => github.com/googleapis/go-genproto v0.0.0-20190819201941-24fa4b261c55
	google.golang.org/grpc => github.com/grpc/grpc-go v1.23.0
)

go 1.13
