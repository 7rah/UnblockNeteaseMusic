package config

var (
	Port     = 7777         //flag.Int("p", 80, "specify server port,such as : \"80\"")
	Source   = "kuwo:kugou" //flag.String("o", "kuwo:kugou", "specify server source,such as : \"kuwo:kugou\"")
	Mode     = 0            //flag.Int("m", 1, "specify running mode（1:hosts） ,such as : \"1\"")
	EndPoint = false        //flag.Bool("e", false, "replace song url")
	TLSPort  = 443
)
