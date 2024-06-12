package wallet

type GetBalanceRequest struct {
	Address string `json:"address"`
}

type GetBalanceResponse struct {
	Balance float64 `json:"balance"`
}

type CreateWalletResponse struct {
	PublicKey string   `json:"publicKey"`
	SecretKey string   `json:"secretKey"`
	Mnemonic  []string `json:"mnemonic"`
	Address   string   `json:"address"`
}
