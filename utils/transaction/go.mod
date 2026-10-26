module github.com/ARM-software/golang-utils/utils/transaction

go 1.26.0

require (
	github.com/go-faker/faker/v4 v4.7.0
	github.com/stretchr/testify v1.12.1
	go.uber.org/mock v0.6.0
)

require (
	github.com/ARM-software/golang-utils/utils/collection v1.170.7 // indirect
	github.com/ARM-software/golang-utils/utils/field v1.170.7 // indirect
	github.com/ARM-software/golang-utils/utils/ptr v1.170.7 // indirect
	github.com/ARM-software/golang-utils/utils/safecast v1.170.7 // indirect
	github.com/ARM-software/golang-utils/utils/value v1.170.7 // indirect
	github.com/deckarep/golang-set/v2 v2.9.0 // indirect
	github.com/gofrs/uuid/v5 v5.5.1 // indirect
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/petermattis/goid v0.0.0-20250813065127-a731cc31b4fe // indirect
	github.com/sasha-s/go-deadlock v0.3.9 // indirect
	go.mongodb.org/mongo-driver v1.17.4 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
)

require (
	github.com/ARM-software/golang-utils/utils/commonerrors v1.170.7
	github.com/ARM-software/golang-utils/utils/commonerrors/errortest v1.170.7
	github.com/ARM-software/golang-utils/utils/idgen v1.170.7
	github.com/ARM-software/golang-utils/utils/parallelisation v1.170.7
	github.com/ARM-software/golang-utils/utils/reflection v1.170.7
	golang.org/x/mod v0.41.0 // indirect
	golang.org/x/sync v0.24.0 // indirect
	golang.org/x/text v0.43.0 // indirect
	golang.org/x/tools v0.51.0 // indirect
)

tool go.uber.org/mock/mockgen

replace github.com/ARM-software/golang-utils/utils/commonerrors => ../commonerrors

replace github.com/ARM-software/golang-utils/utils/commonerrors/errortest => ../commonerrors/errortest

replace github.com/ARM-software/golang-utils/utils/idgen => ../idgen

replace github.com/ARM-software/golang-utils/utils/parallelisation => ../parallelisation

replace github.com/ARM-software/golang-utils/utils/reflection => ../reflection

replace github.com/ARM-software/golang-utils/utils => ../../utils

replace github.com/ARM-software/golang-utils/utils/cache => ../cache

replace github.com/ARM-software/golang-utils/utils/charset => ../charset

replace github.com/ARM-software/golang-utils/utils/collection => ../collection

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

replace github.com/ARM-software/golang-utils/utils/keyring => ../keyring

replace github.com/ARM-software/golang-utils/utils/licensing => ../licensing

replace github.com/ARM-software/golang-utils/utils/logs => ../logs

replace github.com/ARM-software/golang-utils/utils/logs/logstest => ../logs/logstest

replace github.com/ARM-software/golang-utils/utils/maps => ../maps

replace github.com/ARM-software/golang-utils/utils/mocks => ../mocks

replace github.com/ARM-software/golang-utils/utils/parallelisation/mocks => ../parallelisation/mocks

replace github.com/ARM-software/golang-utils/utils/platform => ../platform

replace github.com/ARM-software/golang-utils/utils/proc => ../proc

replace github.com/ARM-software/golang-utils/utils/ptr => ../ptr

replace github.com/ARM-software/golang-utils/utils/resource => ../resource

replace github.com/ARM-software/golang-utils/utils/retry => ../retry

replace github.com/ARM-software/golang-utils/utils/safecast => ../safecast

replace github.com/ARM-software/golang-utils/utils/safeio => ../safeio

replace github.com/ARM-software/golang-utils/utils/semver => ../semver

replace github.com/ARM-software/golang-utils/utils/serialization => ../serialization

replace github.com/ARM-software/golang-utils/utils/serialization/maps/testing => ../serialization/maps/testing

replace github.com/ARM-software/golang-utils/utils/sharedcache => ../sharedcache

replace github.com/ARM-software/golang-utils/utils/signing => ../signing

replace github.com/ARM-software/golang-utils/utils/strings => ../strings

replace github.com/ARM-software/golang-utils/utils/subprocess => ../subprocess

replace github.com/ARM-software/golang-utils/utils/units => ../units

replace github.com/ARM-software/golang-utils/utils/url => ../url

replace github.com/ARM-software/golang-utils/utils/validation => ../validation

replace github.com/ARM-software/golang-utils/utils/value => ../value

replace github.com/ARM-software/golang-utils/utils/serialisation => ../serialisation

replace github.com/ARM-software/golang-utils/utils/serialisation/maps/testing => ../serialisation/maps/testing
