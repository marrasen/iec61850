package iec61850

// #include <iec61850_client.h>
import "C"

import (
	"fmt"
	"log"
	"strings"
)

func sameDataSet(a, b string) bool {
	// normalize LN/DataSet separator: $ or .
	norm := func(s string) string {
		// only replace the *first* '$' after LD/LN if you want to be strict;
		// as a simple first pass, replace all.
		return strings.ReplaceAll(s, "$", ".")
	}
	return norm(a) == norm(b)
}

// PickAndEnableStatDRBRCB selects a free buffered report control block (BRCB)
// under <ld>/<ln> matching the prefix "rcbStatDR" (case-sensitive), configures
// it to the provided datasetRef, sets reasonable trigger options and timing,
// enables reporting, and returns the full RCB reference and a cleanup function
// that disables the RCB (RptEna=false).
//
// The function tries all matching RCBs and returns the first that can be
// successfully enabled. If none can be enabled it returns an error.
func (c *Client) PickAndEnableStatDRBRCB(ld, ln string, datasetRef string) (string, func() error, error) {
	lnRef := fmt.Sprintf("%s/%s", ld, ln)

	log.Printf("Picking free StatDR BRCB for %s", lnRef)

	// List all BRCB names for LLN0
	names, err := c.GetLogicalNodeDirectory(fmt.Sprintf("%s/%s", ld, ln), ACSI_CLASS_BRCB)
	if err != nil {
		return "", nil, fmt.Errorf("failed to list BRCBs for %s: %w", lnRef, err)
	}

	// Try all RCBs with prefix rcbStatDR
	for _, name := range names {
		if !strings.HasPrefix(name, "rcbStatDR") { // exact, case-sensitive prefix match
			log.Printf("Note: Skipping %s (not a rcbStatDRxx)", name)
			continue
		}
		rcbRef := fmt.Sprintf("%s.BR.%s", lnRef, name)

		log.Printf("Trying %s", rcbRef)

		// ReadObject current values
		rcb, err := c.GetRCBValues(rcbRef)
		if err != nil || rcb == nil {
			log.Printf("Note: GetRCBValues for %s failed, will continue to next rcb. Error was: %s", rcbRef, err)
			continue
		}

		if !sameDataSet(rcb.DatSet, datasetRef) {
			// Dataset mismatch - try next candidate
			log.Printf("Note: Dataset mismatch for %s, will continue to next rcb. Expected %s, got %s", rcbRef, datasetRef, rcb.DatSet)
			continue
		}

		log.Printf("Found free StatDR BRCB %s related to DS=%s", rcbRef, datasetRef)

		// If already enabled, try to disable first to take ownership
		if rcb.Ena {
			if err := c.SetRptEna(rcbRef, false); err != nil {
				// cannot disable -> try next
				continue
			}
			// re-read to confirm disabled
			rcb, err = c.GetRCBValues(rcbRef)
			if err != nil {
				log.Printf("Note: RptEna=true for %s, will continue to next rcb. Error was: %s", rcbRef, err)
				continue
			}
			if rcb == nil {
				log.Printf("Note: RptEna=true for %s, will continue to next rcb. Rcb is nil.", rcbRef)
				continue
			}
			if rcb.Ena {
				log.Printf("Note: RptEna=true for %s, will continue to next rcb. RptEna=true.", rcbRef)
				continue
			}
		}

		// Configure trigger options and timing
		ops := TrgOps{DataChange: true, TriggeredPeriodically: false, Gi: true}
		if err := c.SetTrgOps(rcbRef, ops); err != nil {
			// some IEDs may restrict changes, still continue
			log.Printf("Note: SetTrgOps failed for %s, will ignore. Error was: %s", rcbRef, err)
		}
		// BufTm and IntgPd typical for StatDR
		if err := c.SetBufTm(rcbRef, 50); err != nil {
			log.Printf("Note: SetBufTm failed for %s, will ignore. Error was: %s", rcbRef, err)
		}
		/*if err := c.SetIntgPd(rcbRef, 10000); err != nil {
			log.Printf("Note: SetIntgPd failed for %s, will ignore. Error was: %s", rcbRef, err)
		}*/
		if err := c.SetGI(rcbRef, true); err != nil {
			log.Printf("Note: SetGI failed for %s, will ignore. Error was: %s", rcbRef, err)
		}

		// Enable
		if err := c.SetRptEna(rcbRef, true); err != nil {
			log.Printf("Note: SetRptEna failed for %s, will continue to next rcb. Error was: %s", rcbRef, err)
			// try next candidate when enabling fails
			continue
		}
		// Success
		cleanup := func() error {
			return c.SetRptEna(rcbRef, false)
		}
		return rcbRef, cleanup, nil
	}

	return "", nil, fmt.Errorf("no free rcbStatDRxx available under %s", lnRef)
}
