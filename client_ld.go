package iec61850

// #include <iec61850_client.h>
import "C"
import (
	"fmt"
	"log"
	"strings"
	"sync"
	"unsafe"

	"golang.org/x/sync/errgroup"
)

// GetVariableValues is an example of how to read variable structure and values from an IED
func (c *Client) GetVariableValues() ([]VariableTypeValue, error) {
	log.Printf("Loading data model")
	if err := c.GetDeviceModelFromServer(); err != nil {
		return nil, err
	}
	log.Printf("Done")

	// Use Go wrapper to fetch logical device names
	ldNames, err := c.GetLogicalDeviceList()
	if err != nil {
		return nil, err
	}

	ret := make([]VariableTypeValue, 0)
	ch := make(chan []VariableTypeValue)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for v := range ch {
			ret = append(ret, v...)
		}
	}()

	eg := errgroup.Group{}
	eg.SetLimit(2)

	type qvars struct {
		ld   string
		vars []FCVar
	}
	q := make([]qvars, 0)
	totCount := 0

	for _, ldName := range ldNames {
		fmt.Printf("Reading variables for LD %s\n", ldName)
		variables, err := c.GetLogicalDeviceVariablesHierarchical(ldName)
		if err != nil {
			return nil, err
		}
		totCount += len(variables)
		q = append(q, qvars{ldName, variables})
	}
	progress := 0
	for _, qv := range q {
		for _, v := range qv.vars {
			progress++
			fmt.Printf("\rReading object values: %d/%d - %s/%s", progress, totCount, qv.ld, v.LN)

			ldName := qv.ld
			for fc := range v.FCVars {
				eg.Go(func() error {
					dataRef := fmt.Sprintf("%s/%s", ldName, v.LN)

					values, err := c.GetVariableTypeValues(dataRef, FunctionalConstraintFromString(fc))
					if err != nil {
						ch <- []VariableTypeValue{{
							Type:  0,
							Name:  v.LN,
							Ref:   dataRef,
							Value: err,
						}}
						return nil
					}

					//fmt.Printf("  FC: %s - %s - [%s]\n    Value: %s\n", fc, dataRef, strings.Join(subVars, ", "), object)
					ch <- values
					return nil
				})
			}
		}
		fmt.Printf("\n")
	}

	err = eg.Wait()
	close(ch)
	wg.Wait()

	return ret, err
}

func (c *Client) GetDataModel() (DataModel, error) {
	if err := c.GetDeviceModelFromServer(); err != nil {
		return DataModel{}, err
	}

	// Use Go wrapper to fetch logical device names
	ldNames, err := c.GetLogicalDeviceList()
	if err != nil {
		return DataModel{}, err
	}

	var dataModel DataModel
	for _, name := range ldNames {
		var ld LD
		ld.Data = name
		dataModel.LDs = append(dataModel.LDs, ld)
	}

	for i, ld := range dataModel.LDs {
		logicalNodes, err := c.GetLogicalDeviceDirectory(ld.Data)
		if err != nil {
			return DataModel{}, err
		}

		for _, lnName := range logicalNodes {
			var ln LN
			ln.Data = lnName
			lnRef := fmt.Sprintf("%s/%s", ld.Data, lnName)
			ln.Ref = lnRef
			ld.LNs = append(ld.LNs, ln)
		}

		for j, ln := range ld.LNs {
			lnRef := ln.Ref

			dataObjects, err := c.GetLogicalNodeDirectory(lnRef, ACSI_CLASS_DATA_OBJECT)
			if err != nil {
				return DataModel{}, err
			}

			for _, doName := range dataObjects {
				var do DO
				do.Data = doName
				ln.DOs = append(ln.DOs, do)
			}

			for k, do := range ln.DOs {
				doRef := fmt.Sprintf("%s/%s.%s", ld.Data, ln.Data, do.Data)

				ln.DOs[k].DAs, err = c.GetDAs(doRef)
				if err != nil {
					return DataModel{}, err
				}
			}

			// Use wrapper to get logical node directory for DATA_SET
			dataSets, err := c.GetLogicalNodeDirectory(lnRef, ACSI_CLASS_DATA_SET)
			if err != nil {
				return DataModel{}, err
			}
			for _, dsName := range dataSets {
				var ds DS
				ds.Data = dsName
				dataSetRef := fmt.Sprintf("%s.%s", lnRef, ds.Data)
				// Use wrapper to get dataset members and deletable flag
				dataSetMembers, isDeletable, err := c.GetDataSetDirectory(dataSetRef)
				if err != nil {
					return DataModel{}, err
				}
				ds.IsDeletable = isDeletable
				for _, member := range dataSetMembers {
					var dsRef DSRef
					dsRef.Data = member
					ds.DSRefs = append(ds.DSRefs, dsRef)
				}
				ln.DSs = append(ln.DSs, ds)
			}

			// Use wrapper for URCB directory
			reports, err := c.GetLogicalNodeDirectory(lnRef, ACSI_CLASS_URCB)
			if err != nil {
				return DataModel{}, err
			}
			for _, name := range reports {
				var r URReport
				r.Data = name
				r.Ref = fmt.Sprintf("%s.%s", lnRef, r.Data)
				ln.URReports = append(ln.URReports, r)
			}

			// Use wrapper for BRCB directory
			reports, err = c.GetLogicalNodeDirectory(lnRef, ACSI_CLASS_BRCB)
			if err != nil {
				return DataModel{}, err
			}
			for _, name := range reports {
				var r BRReport
				r.Data = name
				r.Ref = fmt.Sprintf("%s.%s", lnRef, r.Data)
				ln.BRReports = append(ln.BRReports, r)
			}

			ld.LNs[j] = ln
		}
		dataModel.LDs[i] = ld
	}
	return dataModel, nil
}

func (c *Client) GetDAs(doRef string) ([]DA, error) {
	// Use Go wrapper to obtain data attribute names (may include FC suffix like "DA1[ST]")
	rawNames, err := c.GetDataDirectoryFC(doRef)
	if err != nil {
		return nil, err
	}

	var das []DA
	for _, rawName := range rawNames {
		var da DA

		// Extract optional FC suffix like "name[ST]"
		name := rawName
		fc := NONE
		if i := strings.LastIndex(rawName, "["); i != -1 && strings.HasSuffix(rawName, "]") && i < len(rawName)-1 {
			fcStr := rawName[i+1 : len(rawName)-1]
			fc = FunctionalConstraintFromString(fcStr)
			// strip suffix from name used for Data/Ref
			name = rawName[:i]
		}

		da.Data = name
		da.FC = fc
		da.Ref = fmt.Sprintf("%s.%s", doRef, da.Data)

		// Recurse for sub DAs using the clean reference (without FC suffix)
		da.DAs, err = c.GetDAs(da.Ref)
		if err != nil {
			return nil, err
		}

		das = append(das, da)
	}

	return das, nil
}

// GetLogicalDeviceDirectory wraps C.IedConnection_getLogicalDeviceDirectory and returns Go strings
func (c *Client) GetLogicalDeviceDirectory(logicalDeviceName string) ([]string, error) {
	var clientError C.IedClientError
	cStr := Go2CStr(logicalDeviceName)
	defer C.free(unsafe.Pointer(cStr))

	list := C.IedConnection_getLogicalDeviceDirectory(c.conn, &clientError, cStr)
	defer func() {
		if list != nil {
			C.LinkedList_destroy(list)
		}
	}()
	if err := GetIedClientError(clientError); err != nil {
		return nil, err
	}
	// convert to []string
	var out []string
	if list != nil {
		it := list.next
		for it != nil {
			out = append(out, C2GoStr((*C.char)(it.data)))
			it = it.next
		}
	}
	return out, nil
}

// GetLogicalNodeDirectory wraps C.IedConnection_getLogicalNodeDirectory and returns Go strings
func (c *Client) GetLogicalNodeDirectory(logicalNodeReference string, acsiClass ACSIClass) ([]string, error) {
	var clientError C.IedClientError
	cRef := Go2CStr(logicalNodeReference)
	defer C.free(unsafe.Pointer(cRef))

	list := C.IedConnection_getLogicalNodeDirectory(c.conn, &clientError, cRef, C.ACSIClass(C.int(acsiClass)))
	defer func() {
		if list != nil {
			C.LinkedList_destroy(list)
		}
	}()
	if err := GetIedClientError(clientError); err != nil {
		return nil, err
	}
	var out []string
	if list != nil {
		it := list.next
		for it != nil {
			out = append(out, C2GoStr((*C.char)(it.data)))
			it = it.next
		}
	}
	return out, nil
}

// GetDataSetDirectory wraps C.IedConnection_getDataSetDirectory and returns Go strings
func (c *Client) GetDataSetDirectory(dataSetReference string) ([]string, bool, error) {
	var clientError C.IedClientError
	var isDeletable C.bool

	cRef := Go2CStr(dataSetReference)
	defer C.free(unsafe.Pointer(cRef))

	list := C.IedConnection_getDataSetDirectory(c.conn, &clientError, cRef, &isDeletable)
	defer func() {
		if list != nil {
			C.LinkedList_destroy(list)
		}
	}()
	if err := GetIedClientError(clientError); err != nil {
		return nil, false, err
	}
	var out []string
	if list != nil {
		it := list.next
		for it != nil {
			out = append(out, C2GoStr((*C.char)(it.data)))
			it = it.next
		}
	}
	return out, bool(isDeletable), nil
}

// GetDataDirectoryFC wraps C.IedConnection_getDataDirectoryFC and returns Go strings
func (c *Client) GetDataDirectoryFC(dataObjectReference string) ([]string, error) {
	var clientError C.IedClientError
	cRef := Go2CStr(dataObjectReference)
	defer C.free(unsafe.Pointer(cRef))

	list := C.IedConnection_getDataDirectoryFC(c.conn, &clientError, cRef)
	defer func() {
		if list != nil {
			C.LinkedList_destroy(list)
		}
	}()
	if err := GetIedClientError(clientError); err != nil {
		return nil, err
	}
	var out []string
	if list != nil {
		it := list.next
		for it != nil {
			out = append(out, C2GoStr((*C.char)(it.data)))
			it = it.next
		}
	}
	return out, nil
}

// GetServerDirectory wraps C.IedConnection_getServerDirectory and returns a Go slice of names.
// When getFileNames is false the list contains logical device names. When true, it would
// contain file names (if implemented by the underlying library per header note).
func (c *Client) GetServerDirectory(getFileNames bool) ([]string, error) {
	var clientError C.IedClientError

	list := C.IedConnection_getServerDirectory(c.conn, &clientError, C.bool(getFileNames))
	defer func() {
		if list != nil {
			C.LinkedList_destroy(list)
		}
	}()

	if err := GetIedClientError(clientError); err != nil {
		return nil, err
	}

	var out []string
	if list != nil {
		it := list.next
		for it != nil {
			out = append(out, C2GoStr((*C.char)(it.data)))
			it = it.next
		}
	}
	return out, nil
}

// GetLogicalDeviceList wraps C.IedConnection_getLogicalDeviceList and returns Go strings.
// This avoids exposing C.LinkedList to callers and centralizes memory management.
func (c *Client) GetLogicalDeviceList() ([]string, error) {
	var clientError C.IedClientError
	list := C.IedConnection_getLogicalDeviceList(c.conn, &clientError)
	defer func() {
		if list != nil {
			C.LinkedList_destroy(list)
		}
	}()
	if err := GetIedClientError(clientError); err != nil {
		return nil, err
	}
	var out []string
	if list != nil {
		it := list.next
		for it != nil {
			out = append(out, C2GoStr((*C.char)(it.data)))
			it = it.next
		}
	}
	return out, nil
}
