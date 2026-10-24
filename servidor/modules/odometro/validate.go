package odometro

import "analisefrota/common/utils"

// valido descarta registro sem placa
func valido(placa string) bool { return utils.ChavePlaca(placa) != "" }
