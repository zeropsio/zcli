package region

type Config struct {
	Name             string `json:"name"`
	IsDefault        bool   `json:"isDefault"`
	RestApiAddress   string `json:"restApiAddress"`
	GrpcApiAddress   string `json:"grpcApiAddress"`
	VpnApiAddress    string `json:"vpnApiAddress"`
	CaCertificateUrl string `json:"caCertificateUrl"`
}
