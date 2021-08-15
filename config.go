package main

type Conf struct {
	Server   Server
	Projects ProjectSettings
	Logging  Logging
}

type Server struct {
	Port      int
	Hostname  string
	UseSSL    bool
	StaticDir string
}

type ProjectSettings struct {
	FileName    string
	MetaRefresh MetaRefresh
	Remote      RemoteProject
}

type MetaRefresh struct {
	Enabled bool
	To      string
}

type RemoteProject struct {
	Enabled          bool
	RefreshFrequency string
	URL              string
}

type Logging struct {
	Method string
	Format string
	Level  string
	File   string
}
