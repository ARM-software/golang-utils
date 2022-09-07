module github.com/ARM-software/golang-utils/utils

go 1.16

require (
	github.com/OneOfOne/xxhash v1.2.8
	github.com/avast/retry-go v3.0.0+incompatible
	github.com/bmatcuk/doublestar/v3 v3.0.0
	github.com/bombsimon/logrusr v1.1.0
	github.com/bxcodec/faker/v3 v3.8.0
	github.com/djherbis/times v1.5.0
	github.com/go-git/go-git/v5 v5.4.2
	github.com/go-http-utils/headers v0.0.0-20181008091004-fed159eddc2a
	github.com/go-logr/logr v0.4.0 // Staying on this version until kubernetes uses a more recent one
	github.com/go-logr/stdr v0.4.0
	github.com/go-ozzo/ozzo-validation/v4 v4.3.0
	github.com/gofrs/uuid v4.2.0+incompatible
	github.com/gogs/chardet v0.0.0-20211120154057-b7413eaefb8f
	github.com/golang/mock v1.6.0
	github.com/hashicorp/go-cleanhttp v0.5.2
	github.com/hashicorp/go-retryablehttp v0.7.1
	github.com/joho/godotenv v1.4.0
	github.com/mitchellh/mapstructure v1.5.0
	github.com/rifflock/lfshook v0.0.0-20180920164130-b9218ef580f5
	github.com/rs/zerolog v1.28.0
	github.com/sasha-s/go-deadlock v0.3.1
	github.com/shirou/gopsutil/v3 v3.22.8
	github.com/sirupsen/logrus v1.9.0
	github.com/spaolacci/murmur3 v1.1.0
	github.com/spf13/afero v1.9.2
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.13.0
	github.com/stretchr/testify v1.8.0
	go.uber.org/atomic v1.9.0
	go.uber.org/goleak v1.1.12
	golang.org/x/net v0.0.0-20220520000938-2e3eb7b945c2
	golang.org/x/sync v0.0.0-20220513210516-0976fa681c29
	golang.org/x/text v0.3.7
)

require golang.org/x/tools v0.1.7 // indirect
