module github.com/ARM-software/golang-utils/utils/platform

go 1.26.0

require (
	github.com/go-faker/faker/v4 v4.7.0
	github.com/go-ozzo/ozzo-validation/v4 v4.4.1
	github.com/iamacarpet/go-win64api v0.0.0-20240507095429-873e84e85847
	github.com/mitchellh/go-homedir v1.1.0
	github.com/shirou/gopsutil/v4 v4.26.5
	github.com/stretchr/testify v1.12.1
	golang.org/x/sys v0.49.0
)

require (
	github.com/ARM-software/golang-utils/utils/encoding v1.170.7 // indirect
	github.com/ARM-software/golang-utils/utils/field v1.170.7 // indirect
	github.com/ARM-software/golang-utils/utils/ptr v1.170.7 // indirect
	github.com/ARM-software/golang-utils/utils/url v1.170.7 // indirect
	github.com/ARM-software/golang-utils/utils/value v1.170.7 // indirect
	github.com/asaskevich/govalidator v0.0.0-20210307081110-f21760c49a8d // indirect
	github.com/deckarep/golang-set/v2 v2.9.0 // indirect
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/petermattis/goid v0.0.0-20250813065127-a731cc31b4fe // indirect
	github.com/sasha-s/go-deadlock v0.3.9 // indirect
	go.mongodb.org/mongo-driver v1.17.4 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/net v0.61.0 // indirect
	golang.org/x/sync v0.24.0 // indirect
)

require (
	github.com/ARM-software/golang-utils/utils/collection v1.170.7
	github.com/ARM-software/golang-utils/utils/commonerrors v1.170.7
	github.com/ARM-software/golang-utils/utils/commonerrors/errortest v1.170.7
	github.com/ARM-software/golang-utils/utils/parallelisation v1.170.7
	github.com/ARM-software/golang-utils/utils/proc v1.170.7
	github.com/ARM-software/golang-utils/utils/reflection v1.170.7
	github.com/ARM-software/golang-utils/utils/safecast v1.170.7
	github.com/ARM-software/golang-utils/utils/subprocess v1.170.7
	github.com/ARM-software/golang-utils/utils/validation v1.170.7
	github.com/ebitengine/purego v0.10.0 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/google/cabbie v1.0.2 // indirect
	github.com/google/glazier v0.0.0-20211029225403-9f766cca891d // indirect
	github.com/lufia/plan9stats v0.0.0-20211012122336-39d0f177ccd0 // indirect
	github.com/power-devops/perfstat v0.0.0-20240221224432-82ca36839d55 // indirect
	github.com/scjalliance/comshim v0.0.0-20190308082608-cf06d2532c4e // indirect
	github.com/tklauser/go-sysconf v0.3.16 // indirect
	github.com/tklauser/numcpus v0.11.0 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	golang.org/x/text v0.43.0 // indirect
)

replace github.com/ARM-software/golang-utils/utils/collection => ../collection

replace github.com/ARM-software/golang-utils/utils/commonerrors => ../commonerrors

replace github.com/ARM-software/golang-utils/utils/commonerrors/errortest => ../commonerrors/errortest

replace github.com/ARM-software/golang-utils/utils/parallelisation => ../parallelisation

replace github.com/ARM-software/golang-utils/utils/proc => ../proc

replace github.com/ARM-software/golang-utils/utils/reflection => ../reflection

replace github.com/ARM-software/golang-utils/utils/safecast => ../safecast

replace github.com/ARM-software/golang-utils/utils/subprocess => ../subprocess

replace github.com/ARM-software/golang-utils/utils/validation => ../validation

replace github.com/ARM-software/golang-utils/utils => ../../utils

replace github.com/ARM-software/golang-utils/utils/cache => ../cache

replace github.com/ARM-software/golang-utils/utils/charset => ../charset

replace github.com/ARM-software/golang-utils/utils/config => ../config

replace github.com/ARM-software/golang-utils/utils/diodes => ../diodes

replace github.com/ARM-software/golang-utils/utils/encoding => ../encoding

replace github.com/ARM-software/golang-utils/utils/encryption => ../encryption

replace github.com/ARM-software/golang-utils/utils/encryption/aesrsa/testhelpers => ../encryption/aesrsa/testhelpers

replace github.com/ARM-software/golang-utils/utils/environment => ../environment

replace github.com/ARM-software/golang-utils/utils/field => ../field

replace github.com/ARM-software/golang-utils/utils/filesystem => ../filesystem

replace github.com/ARM-software/golang-utils/utils/filesystem/filesystemtest => ../filesystem/filesystemtest

replace github.com/ARM-software/golang-utils/utils/git => ../git

replace github.com/ARM-software/golang-utils/utils/hashing => ../hashing

replace github.com/ARM-software/golang-utils/utils/http => ../http

replace github.com/ARM-software/golang-utils/utils/http/httptest => ../http/httptest

replace github.com/ARM-software/golang-utils/utils/idgen => ../idgen

replace github.com/ARM-software/golang-utils/utils/keyring => ../keyring

replace github.com/ARM-software/golang-utils/utils/licensing => ../licensing

replace github.com/ARM-software/golang-utils/utils/logs => ../logs

replace github.com/ARM-software/golang-utils/utils/logs/logstest => ../logs/logstest

replace github.com/ARM-software/golang-utils/utils/maps => ../maps

replace github.com/ARM-software/golang-utils/utils/mocks => ../mocks

replace github.com/ARM-software/golang-utils/utils/parallelisation/mocks => ../parallelisation/mocks

replace github.com/ARM-software/golang-utils/utils/ptr => ../ptr

replace github.com/ARM-software/golang-utils/utils/resource => ../resource

replace github.com/ARM-software/golang-utils/utils/retry => ../retry

replace github.com/ARM-software/golang-utils/utils/safeio => ../safeio

replace github.com/ARM-software/golang-utils/utils/semver => ../semver

replace github.com/ARM-software/golang-utils/utils/serialization => ../serialization

replace github.com/ARM-software/golang-utils/utils/serialization/maps/testing => ../serialization/maps/testing

replace github.com/ARM-software/golang-utils/utils/sharedcache => ../sharedcache

replace github.com/ARM-software/golang-utils/utils/signing => ../signing

replace github.com/ARM-software/golang-utils/utils/strings => ../strings

replace github.com/ARM-software/golang-utils/utils/transaction => ../transaction

replace github.com/ARM-software/golang-utils/utils/units => ../units

replace github.com/ARM-software/golang-utils/utils/url => ../url

replace github.com/ARM-software/golang-utils/utils/value => ../value
