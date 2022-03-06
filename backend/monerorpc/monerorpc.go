package monerorpc

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/gorilla/rpc/v2/json2"
)

type (
	Config struct {
		Address       string
		CustomHeaders map[string]string
		Transport     http.RoundTripper
	}
	Client struct {
		httpcl  *http.Client
		addr    string
		headers map[string]string
	}

	Address struct {
		AddressIndex      uint64 `json:"address_index"`
		Address           string `json:"address"`
		Balance           uint64 `json:"balance"`
		UnlockedBalance   uint64 `json:"unlocked_balance"`
		Label             string `json:"label"`
		Tag               string `json:"tag"`
		NumUnspentOutputs uint64 `json:"num_unspent_outputs"`
		Used              bool   `json:"used"`
	}

	SubaddressIndex struct {
		Major uint64 `json:"major"`
		Minor uint64 `json:"minor"`
	}

	Destination struct {
		Amount  uint64 `json:"amount"`
		Address string `json:"address"`
	}

	Payment struct {
		PaymentId    string          `json:"payment_id"`
		TxHash       string          `json:"tx_hash"`
		Amount       uint64          `json:"amount"`
		BlockHeight  uint64          `json:"block_height"`
		UnlockTime   uint64          `json:"unlock_time"`
		SubaddrIndex SubaddressIndex `json:"subaddr_index"`
		Major        uint64          `json:"major"`
		Minor        uint64          `json:"minor"`
		Address      string          `json:"address"`
	}

	SignedKeyImage struct {
		KeyImage  string `json:"key_image"`
		Signature string `json:"signature"`
	}

	Entry struct {
		Address     string `json:"address"`
		Description string `json:"description"`
		Index       uint64 `json:"index"`
	}

	ValidateAddressRequest struct {
		Address        string `json:"address"`
		AnyNetType     bool   `json:"any_net_type,omitempty"`
		AllowOpenalias bool   `json:"allow_openalias,omitempty"`
	}

	ValidateAddressResponse struct {
		Valid      bool   `json:"valid"`
		Integrated bool   `json:"integrated"`
		Subaddress bool   `json:"subaddress"`
		Nettype    string `json:"nettype"`
	}

	CreateAddressRequest struct {
		AccountIndex uint64 `json:"account_index"`
		Label        string `json:"label,omitempty"`
	}

	CreateAddressResponse struct {
		Address      string `json:"address"`
		AddressIndex uint64 `json:"address_index"`
	}
	GetBalanceRequest struct {
		AccountIndex   uint64   `json:"account_index"`
		AddressIndices []uint64 `json:"address_indices,omitempty"`
	}

	GetBalanceResponse struct {
		Balance              uint64    `json:"balance"`
		UnlockedBalance      uint64    `json:"unlocked_balance"`
		MultisigImportNeeded bool      `json:"multisig_import_needed"`
		PerSubaddress        []Address `json:"per_subaddress"`
	}

	GetTransfersRequest struct {
		In             bool     `json:"in,omitempty"`
		Out            bool     `json:"out,omitempty"`
		Pending        bool     `json:"pending,omitempty"`
		Failed         bool     `json:"failed,omitempty"`
		Pool           bool     `json:"pool,omitempty"`
		FilterByHeight bool     `json:"filter_by_height,omitempty"`
		MinHeight      uint64   `json:"min_height,omitempty"`
		MaxHeight      uint64   `json:"max_height,omitempty"`
		AccountIndex   uint64   `json:"account_index,omitempty"`
		SubaddrIndices []uint64 `json:"subaddr_indices,omitempty"`
	}

	Transfer struct {
		Address                         string          `json:"address"`
		Amount                          uint64          `json:"amount"`
		Confirmations                   uint64          `json:"confirmations"`
		DoubleSpendSeen                 bool            `json:"double_spend_seen"`
		Fee                             uint64          `json:"fee"`
		Height                          uint64          `json:"height"`
		Note                            string          `json:"note"`
		PaymentId                       string          `json:"payment_id"`
		SubaddrIndex                    SubaddressIndex `json:"subaddr_index"`
		SuggestedConfirmationsThreshold uint64          `json:"suggested_confirmations_threshold"`
		Timestamp                       uint64          `json:"timestamp"`
		Txid                            string          `json:"txid"`
		Type                            string          `json:"type"`
		UnlockTime                      uint64          `json:"unlock_time"`
	}

	GetTransfersResponse struct {
		In      []Transfer `json:"in"`
		Out     []Transfer `json:"out"`
		Pending []Transfer `json:"pending"`
		Failed  []Transfer `json:"failed"`
		Pool    []Transfer `json:"pool"`
	}

	TransferRequest struct {
		Destinations   []Destination `json:"destinations"`
		AccountIndex   uint64        `json:"account_index,omitempty"`
		SubaddrIndices []uint64      `json:"subaddr_indices,omitempty"`
		Mixin          uint64        `json:"mixin"`
		RingSize       uint64        `json:"ring_size"`
		UnlockTime     uint64        `json:"unlock_time"`
		GetTxKeys      bool          `json:"get_tx_keys,omitempty"`
		DoNotRelay     bool          `json:"do_not_relay,omitempty"`
		GetTxHex       bool          `json:"get_tx_hex"`
		GetTxMetadata  bool          `json:"get_tx_metadata"`
	}

	TransferResponse struct {
		TxHash        string `json:"tx_hash"`
		TxKey         string `json:"tx_key"`
		Amount        uint64 `json:"amount"`
		Fee           uint64 `json:"fee"`
		TxBlob        string `json:"tx_blob"`
		TxMetadata    string `json:"tx_metadata"`
		UnsignedTxset string `json:"unsigned_txset"`
	}

	TransferSplitRequest struct {
		Destinations   []Destination `json:"destinations"`
		AccountIndex   uint64        `json:"account_index,omitempty"`
		SubaddrIndices []uint64      `json:"subaddr_indices,omitempty"`
		Mixin          uint64        `json:"mixin"`
		RingSize       uint64        `json:"ring_size"`
		UnlockTime     uint64        `json:"unlock_time"`
		GetTxKeys      bool          `json:"get_tx_keys,omitempty"`
		Priority       uint64        `json:"priority"`
		DoNotRelay     bool          `json:"do_not_relay,omitempty"`
		GetTxHex       bool          `json:"get_tx_hex"`
		NewAlgorithm   bool          `json:"new_algorithm"`
		GetTxMetadata  bool          `json:"get_tx_metadata"`
	}

	TransferSplitResponse struct {
		TxHashList     []string `json:"tx_hash_list"`
		TxKeyList      []string `json:"tx_key_list"`
		AmountList     []int    `json:"amount_list"`
		FeeList        []int    `json:"fee_list"`
		TxBlobList     []string `json:"tx_blob_list"`
		TxMetadataList []string `json:"tx_metadata_list"`
		MultisigTxset  string   `json:"multisig_txset"`
		UnsignedTxset  string   `json:"unsigned_txset"`
	}

	MakeUriRequest struct {
		Address       string `json:"address"`
		Amount        uint64 `json:"amount,omitempty"`
		PaymentId     string `json:"payment_id,omitempty"`
		RecipientName string `json:"recipient_name,omitempty"`
		TxDescription string `json:"tx_description,omitempty"`
	}

	MakeUriResponse struct {
		Uri string `json:"uri"`
	}

	GetAddressRequest struct {
		AccountIndex uint64   `json:"account_index"`
		AddressIndex []uint64 `json:"address_index,omitempty"`
	}

	GetAddressResponse struct {
		Address   string    `json:"address"`
		Addresses []Address `json:"addresses"`
	}

	BlockHeader struct {
		Hash      string `json:"hash"`
		Height    uint64 `json:"height"`
		Timestamp uint64 `json:"timestamp"`
	}

	GetBlockHeadersRangeRequest struct {
		StartHeight uint64 `json:"start_height"`
		EndHeight   uint64 `json:"end_height"`
	}

	GetLastBlockHeaderResponse struct {
		BlockHeader BlockHeader `json:"block_header"`
	}

	GetLastBlockHeadersRangeResponse struct {
		BlockHeaders []BlockHeader `json:"headers"`
	}
)

var (
	fakeResponse map[string]func(interface{}) string
)

func SetFakeResponse(method string, cb func(interface{}) string) {
	if fakeResponse == nil {
		fakeResponse = make(map[string]func(interface{}) string)
	}
	fakeResponse[method] = cb
}

func New(cfg Config) *Client {
	cl := &Client{
		addr:    cfg.Address,
		headers: cfg.CustomHeaders,
	}
	if cfg.Transport == nil {
		cl.httpcl = http.DefaultClient
	} else {
		cl.httpcl = &http.Client{
			Transport: cfg.Transport,
		}
	}
	return cl
}

func (c *Client) Do(method string, in, out interface{}) error {
	if fakeResponse != nil {
		if cb, ok := fakeResponse[method]; ok {
			body := fmt.Sprintf(`{"result":%s}`, cb(in))
			return json2.DecodeClientResponse(bytes.NewReader([]byte(body)), out)
		}
		return fmt.Errorf("method: %s not handled in: %v out: %v", method, in, out)
	}
	payload, err := json2.EncodeClientRequest(method, in)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, c.addr, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.headers != nil {
		for k, v := range c.headers {
			req.Header.Set(k, v)
		}
	}
	resp, err := c.httpcl.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("http status %v", resp.StatusCode)
	}
	defer resp.Body.Close()

	if out == nil {
		v := &json2.EmptyResponse{}
		return json2.DecodeClientResponse(resp.Body, v)
	}
	return json2.DecodeClientResponse(resp.Body, out)
}

func (c *Client) ValidateAddress(req *ValidateAddressRequest) (*ValidateAddressResponse, error) {
	resp := &ValidateAddressResponse{}
	err := c.Do("validate_address", &req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) CreateAddress(req *CreateAddressRequest) (*CreateAddressResponse, error) {
	resp := &CreateAddressResponse{}
	err := c.Do("create_address", &req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) GetBalance(req *GetBalanceRequest) (*GetBalanceResponse, error) {
	resp := &GetBalanceResponse{}
	err := c.Do("get_balance", &req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) GetTransfers(req *GetTransfersRequest) (*GetTransfersResponse, error) {
	resp := &GetTransfersResponse{}
	err := c.Do("get_transfers", &req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) Transfer(req *TransferRequest) (*TransferResponse, error) {
	resp := &TransferResponse{}
	err := c.Do("transfer", &req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) TransferSplit(req *TransferSplitRequest) (*TransferSplitResponse, error) {
	resp := &TransferSplitResponse{}
	err := c.Do("transfer_split", &req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func XMRToDecimal(xmr uint64) string {
	str0 := fmt.Sprintf("%013d", xmr)
	l := len(str0)
	return str0[:l-12] + "." + str0[l-12:]
}

func XMRToFloat64(xmr uint64) float64 {
	return float64(xmr) / 1e12
}

func (c *Client) MakeUri(req *MakeUriRequest) (*MakeUriResponse, error) {
	resp := &MakeUriResponse{}
	err := c.Do("make_uri", &req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) GetAddress(req *GetAddressRequest) (*GetAddressResponse, error) {
	resp := &GetAddressResponse{}
	err := c.Do("get_address", &req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) GetLastBlockHeader() (*GetLastBlockHeaderResponse, error) {
	resp := &GetLastBlockHeaderResponse{}
	err := c.Do("get_last_block_header", nil, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) GetBlockHeadersRange(req *GetBlockHeadersRangeRequest) (*GetLastBlockHeadersRangeResponse, error) {
	resp := &GetLastBlockHeadersRangeResponse{}
	err := c.Do("get_block_headers_range", &req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
