package handler

import (
	"concert-go/internal/config"
	"concert-go/internal/util"

	"github.com/gin-gonic/gin"
)

type DevHandler struct {
	cfg *config.Config
}

func NewDevHandler(cfg *config.Config) *DevHandler {
	return &DevHandler{cfg: cfg}
}

type HashPasswordRequest struct {
	Password string `json:"password" binding:"required"`
}

type HashPasswordResponse struct {
	Hash string `json:"hash"`
}

type AESEncryptRequest struct {
	Plaintext string `json:"plaintext" binding:"required"`
}

type AESEncryptResponse struct {
	Ciphertext string `json:"ciphertext"`
}

type AESDecryptRequest struct {
	Ciphertext string `json:"ciphertext" binding:"required"`
}

type AESDecryptResponse struct {
	Plaintext string `json:"plaintext"`
}

// HashPassword godoc
// @Summary Hash a password (DEV ONLY - DELETE BEFORE MAIN)
// @Description Returns bcrypt hash for testing. Remove in production.
// @Tags dev
// @Accept json
// @Produce json
// @Param request body HashPasswordRequest true "Password to hash"
// @Success 200 {object} HashPasswordResponse
// @Router /dev/hash-password [post]
func (h *DevHandler) HashPassword(c *gin.Context) {
	var req HashPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.RespondBadRequest(c, "Invalid request", err.Error())
		return
	}

	hash, err := util.HashPassword(req.Password)
	if err != nil {
		util.RespondInternalError(c)
		return
	}

	util.RespondOK(c, "Password hashed", HashPasswordResponse{Hash: hash}, nil)
}

// AESEncrypt godoc
// @Summary Encrypt plaintext with AES-256-GCM (DEV ONLY - DELETE BEFORE MAIN)
// @Description Returns hex-encoded ciphertext for testing. Remove in production.
// @Tags dev
// @Accept json
// @Produce json
// @Param request body AESEncryptRequest true "Plaintext to encrypt"
// @Success 200 {object} AESEncryptResponse
// @Router /dev/aes-encrypt [post]
func (h *DevHandler) AESEncrypt(c *gin.Context) {
	var req AESEncryptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.RespondBadRequest(c, "Invalid request", err.Error())
		return
	}

	ciphertext, err := util.EncryptAES(h.cfg.AESSecret, req.Plaintext)
	if err != nil {
		util.RespondBadRequest(c, "Encryption failed", err.Error())
		return
	}

	util.RespondOK(c, "Encrypted", AESEncryptResponse{Ciphertext: ciphertext}, nil)
}

// AESDecrypt godoc
// @Summary Decrypt AES-256-GCM ciphertext (DEV ONLY - DELETE BEFORE MAIN)
// @Description Returns plaintext for testing. Remove in production.
// @Tags dev
// @Accept json
// @Produce json
// @Param request body AESDecryptRequest true "Ciphertext to decrypt"
// @Success 200 {object} AESDecryptResponse
// @Router /dev/aes-decrypt [post]
func (h *DevHandler) AESDecrypt(c *gin.Context) {
	var req AESDecryptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.RespondBadRequest(c, "Invalid request", err.Error())
		return
	}

	plaintext, err := util.DecryptAES(h.cfg.AESSecret, req.Ciphertext)
	if err != nil {
		util.RespondBadRequest(c, "Decryption failed", err.Error())
		return
	}

	util.RespondOK(c, "Decrypted", AESDecryptResponse{Plaintext: plaintext}, nil)
}