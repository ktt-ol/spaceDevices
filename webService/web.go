package webService

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/ktt-ol/spaceDevices/conf"
	"github.com/ktt-ol/spaceDevices/db"
	"github.com/ktt-ol/spaceDevices/mqtt"
	"github.com/sirupsen/logrus"
)

var logger = logrus.WithField("where", "webSrv")

var devices *mqtt.DeviceData
var macDb db.UserDb
var xsrfCheck *SimpleXSRFCheck

func StartWebService(conf conf.ServerConf, _devices *mqtt.DeviceData, _macDb db.UserDb) {
	devices = _devices
	macDb = _macDb
	xsrfCheck = NewSimpleXSRFCheck()

	// use logrus logging
	gin.DisableConsoleColor()
	gin.DefaultWriter = logrus.WithField("where", "gin").WriterLevel(logrus.DebugLevel)
	gin.DefaultErrorWriter = logrus.WithField("where", "gin").WriterLevel(logrus.ErrorLevel)

	router := gin.Default()
	router.Use(gzip.Gzip(gzip.DefaultCompression))

	router.Static("/assets", "webUI/assets")
	router.LoadHTMLGlob("webUI/templates/*.html")
	router.GET("/", overviewPageHandler)
	router.POST("/", changeInfoHandler)
	router.GET("/help.html", func(c *gin.Context) {
		c.HTML(http.StatusOK, "help.html", gin.H{})
	})

	addr := fmt.Sprintf("%s:%d", conf.Host, conf.Port)
	var err error
	if conf.Https {
		err = router.RunTLS(addr, conf.CertFile, conf.KeyFile)
	} else {
		err = router.Run(addr)
	}
	if err != nil {
		logger.Error("gin exit", err)
	}
}

func sendError(c *gin.Context, msg string) {
	c.String(http.StatusBadRequest, "Error: "+msg)
	c.Abort()
}

func overviewPageHandler(c *gin.Context) {
	ip, _, _ := net.SplitHostPort(c.Request.RemoteAddr)
	logger.WithField("ip", ip).Debug("Request ip.")

	name := "???"
	mac := "???"
	deviceName := ""
	visibility := db.Visibility(99)
	isLocallyAdministered := false
	macNotFound := false
	if info, ok := devices.GetByIp(ip); ok {
		mac = info.Mac
		isLocallyAdministered = db.IsMacLocallyAdministered(mac)
		if userInfo, ok := macDb.Get(info.Mac); ok {
			name = userInfo.Name
			deviceName = userInfo.DeviceName
			visibility = userInfo.Visibility
		}
	} else {
		macNotFound = true
	}

	c.HTML(http.StatusOK, "index.html", gin.H{
		"secToken":              xsrfCheck.NewToken(ip),
		"name":                  name,
		"mac":                   mac,
		"deviceName":            deviceName,
		"visibility":            visibility,
		"isLocallyAdministered": isLocallyAdministered,
		"macNotFound":           macNotFound,
	})
}

type changeData struct {
	Action     string        `form:"action" binding:"required"`
	SecToken   string        `form:"secToken" binding:"required"`
	Name       string        `form:"name" binding:"required"`
	DeviceName string        `form:"deviceName"`
	Visibility db.Visibility `form:"visibility" binding:"required"`
}

func changeInfoHandler(c *gin.Context) {
	ip, _, _ := net.SplitHostPort(c.Request.RemoteAddr)
	info, ok := devices.GetByIp(ip)
	if !ok {
		logger.WithField("ip", ip).Error("No data for ip found.")
		sendError(c, "No data for your ip found.")
		return
	}

	logger = logger.WithField("mac", info.Mac)

	var form changeData
	if err := c.Bind(&form); err != nil {
		logger.WithError(err).Error("Invalid binding.")
		sendError(c, "Invalid binding.")
		return
	}

	if !xsrfCheck.CheckAndClearToken(ip, form.SecToken) {
		logger.WithFields(logrus.Fields{"ip": ip, "secToken": form.SecToken}).Error("Invalid secToken")
		sendError(c, "Invalid secToken")
		return
	}

	if form.Action == "delete" {
		logger.WithField("user", form.Name).Info("Delete user info.")

		macDb.Delete(info.Mac)
	} else if form.Action == "update" {
		logger.WithField("data", fmt.Sprintf("%#v", form)).Info("Change user info.")

		// visibility, ok := db.ParseVisibility(form.VisibilityNum)
		// if !ok {
		// 	logger.WithField("VisibilityNum", form.VisibilityNum).Error("Invalid visibility.")
		// 	sendError(c, "Invalid 'visibility' value")
		// 	return
		// }

		entry := db.UserDbEntry{Name: form.Name, DeviceName: form.DeviceName, Visibility: form.Visibility, Ts: time.Now().Unix() * 1000}
		macDb.Set(info.Mac, entry)
	}

	c.Redirect(http.StatusSeeOther, "/")
}
