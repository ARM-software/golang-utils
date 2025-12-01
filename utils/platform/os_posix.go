//go:build linux || unix || (js && wasm) || darwin || aix || dragonfly || freebsd || nacl || netbsd || openbsd || solaris

package platform

func expandFromEnvironment(s string) string {
	// nothing to do on unix system as it should be covered by os.ExpandEnv
	return s
}
