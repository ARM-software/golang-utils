# Safecast

the purpose of this utilities is to perform safe number conversion in go similarly to [go-safecast](https://github.com/ccoVeille/go-safecast) from which they are inspired from.
It should help tackling gosec [G115 rule](https://github.com/securego/gosec/pull/1149)

    G115: Potential overflow when converting between integer types.

 and [CWE-190](https://cwe.mitre.org/data/definitions/190.html) 


    infinite loop
    access to wrong resource by id
    grant access to someone who exhausted their quota

Contrary to `go-safecast` no error is returned when attempting casting and the MAX or MIN value of the type is returned instead if the value is beyond the allowed window.
For instance, `toInt8(255)-> 127` and `toInt8(-255)-> -128`



