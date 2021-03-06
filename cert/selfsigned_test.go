package cert_test

import (
	"io/ioutil"
	"testing"

	"github.com/lightningnetwork/lnd/cert"
)

var (
	extraIPs     = []string{"1.1.1.1", "123.123.123.1", "199.189.12.12"}
	extraDomains = []string{"home", "and", "away"}
)

// TestIsOutdatedCert checks that we'll consider the TLS certificate outdated
// if the ip addresses or dns names don't match.
func TestIsOutdatedCert(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "certtest")
	if err != nil {
		t.Fatal(err)
	}

	certPath := tempDir + "/tls.cert"
	keyPath := tempDir + "/tls.key"

	// Generate TLS files with two extra IPs and domains.
	err = cert.GenCertPair(
		"lnd autogenerated cert", certPath, keyPath, extraIPs[:2],
		extraDomains[:2], cert.DefaultAutogenValidity,
	)
	if err != nil {
		t.Fatal(err)
	}

	// We'll attempt to check up-to-date status for all variants of 1-3
	// number of IPs and domains.
	for numIPs := 1; numIPs <= len(extraIPs); numIPs++ {
		for numDomains := 1; numDomains <= len(extraDomains); numDomains++ {
			_, parsedCert, err := cert.LoadCert(
				certPath, keyPath,
			)
			if err != nil {
				t.Fatal(err)
			}

			// Using the test case's number of IPs and domains, get
			// the outdated status of the certificate we created
			// above.
			outdated, err := cert.IsOutdated(
				parsedCert, extraIPs[:numIPs],
				extraDomains[:numDomains],
			)
			if err != nil {
				t.Fatal(err)
			}

			// We expect it to be considered outdated if the IPs or
			// domains don't match exactly what we created.
			expected := numIPs != 2 || numDomains != 2
			if outdated != expected {
				t.Fatalf("expected certificate to be "+
					"outdated=%v, got=%v", expected,
					outdated)
			}
		}
	}
}

// TestIsOutdatedPermutation tests that the order of listed IPs or DNS names,
// nor dulicates in the lists, matter for whether we consider the certificate
// outdated.
func TestIsOutdatedPermutation(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "certtest")
	if err != nil {
		t.Fatal(err)
	}

	certPath := tempDir + "/tls.cert"
	keyPath := tempDir + "/tls.key"

	// Generate TLS files from the IPs and domains.
	err = cert.GenCertPair(
		"lnd autogenerated cert", certPath, keyPath, extraIPs[:],
		extraDomains[:], cert.DefaultAutogenValidity,
	)
	if err != nil {
		t.Fatal(err)
	}
	_, parsedCert, err := cert.LoadCert(certPath, keyPath)
	if err != nil {
		t.Fatal(err)
	}

	// If we have duplicate IPs or DNS names listed, that shouldn't matter.
	dupIPs := make([]string, len(extraIPs)*2)
	for i := range dupIPs {
		dupIPs[i] = extraIPs[i/2]
	}

	dupDNS := make([]string, len(extraDomains)*2)
	for i := range dupDNS {
		dupDNS[i] = extraDomains[i/2]
	}

	outdated, err := cert.IsOutdated(parsedCert, dupIPs, dupDNS)
	if err != nil {
		t.Fatal(err)
	}

	if outdated {
		t.Fatalf("did not expect duplicate IPs or DNS names be " +
			"considered outdated")
	}

	// Similarly, the order of the lists shouldn't matter.
	revIPs := make([]string, len(extraIPs))
	for i := range revIPs {
		revIPs[i] = extraIPs[len(extraIPs)-1-i]
	}

	revDNS := make([]string, len(extraDomains))
	for i := range revDNS {
		revDNS[i] = extraDomains[len(extraDomains)-1-i]
	}

	outdated, err = cert.IsOutdated(parsedCert, revIPs, revDNS)
	if err != nil {
		t.Fatal(err)
	}

	if outdated {
		t.Fatalf("did not expect reversed IPs or DNS names be " +
			"considered outdated")
	}
}
