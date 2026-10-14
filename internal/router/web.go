package router

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// MountFrontend 挂载三端构建产物：
//   /        → admin（登录页）
//   /pos/    → 收银
//   /kds/    → 厨显
// appsRoot 一般为 "web/apps"（相对进程工作目录）。
func MountFrontend(r *gin.Engine, appsRoot string) {
	adminDist := filepath.Join(appsRoot, "admin", "dist")
	posDist := filepath.Join(appsRoot, "pos", "dist")
	kdsDist := filepath.Join(appsRoot, "kds", "dist")

	adminOK := dirExists(adminDist)
	if !adminOK {
		r.GET("/", func(c *gin.Context) {
			c.Header("Content-Type", "text/html; charset=utf-8")
			c.String(http.StatusOK, fallbackHTML())
		})
	} else {
		mountSPA(r, "", adminDist)
	}
	if dirExists(posDist) {
		mountSPA(r, "/pos", posDist)
	}
	if dirExists(kdsDist) {
		mountSPA(r, "/kds", kdsDist)
	}

	r.NoRoute(func(c *gin.Context) {
		p := c.Request.URL.Path
		if strings.HasPrefix(p, "/apis") {
			c.JSON(http.StatusOK, gin.H{"code": 404, "msg": "接口不存在", "data": nil})
			return
		}
		if adminOK && !strings.HasPrefix(p, "/pos") && !strings.HasPrefix(p, "/kds") {
			c.File(filepath.Join(adminDist, "index.html"))
			return
		}
		if strings.HasPrefix(p, "/pos") && dirExists(posDist) {
			c.File(filepath.Join(posDist, "index.html"))
			return
		}
		if strings.HasPrefix(p, "/kds") && dirExists(kdsDist) {
			c.File(filepath.Join(kdsDist, "index.html"))
			return
		}
		c.String(http.StatusNotFound, "404 page not found")
	})
}

func mountSPA(r *gin.Engine, urlPrefix, distDir string) {
	assetsURL := "/assets"
	indexURL := "/"
	if urlPrefix != "" {
		assetsURL = urlPrefix + "/assets"
		indexURL = urlPrefix + "/"
		r.GET(urlPrefix, func(c *gin.Context) {
			c.Redirect(http.StatusFound, indexURL)
		})
	}
	r.Static(assetsURL, filepath.Join(distDir, "assets"))
	r.GET(indexURL, func(c *gin.Context) {
		c.File(filepath.Join(distDir, "index.html"))
	})
}

func dirExists(p string) bool {
	st, err := os.Stat(p)
	return err == nil && st.IsDir()
}

func fallbackHTML() string {
	return `<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8"/>
  <title>Openbows-Rms</title>
  <style>
    body{font-family:system-ui,sans-serif;max-width:640px;margin:48px auto;padding:0 16px;line-height:1.6;color:#222}
    code{background:#f4f4f4;padding:2px 6px;border-radius:4px}
    a{color:#1677ff}
  </style>
</head>
<body>
  <h1>Openbows-Rms API 已启动</h1>
  <p>尚未找到前端构建产物（<code>web/apps/admin/dist</code>）。请在仓库根目录执行：</p>
  <pre>cd web
npm install
npm run build</pre>
  <p>然后重启后端，再访问本页即可打开管理后台登录。</p>
  <p>开发模式也可分别启动：</p>
  <ul>
    <li>管理后台 <code>npm run dev:admin</code> → <a href="http://127.0.0.1:5173">http://127.0.0.1:5173</a></li>
    <li>收银 POS <code>npm run dev:pos</code> → <a href="http://127.0.0.1:5174">http://127.0.0.1:5174</a></li>
    <li>厨显 KDS <code>npm run dev:kds</code> → <a href="http://127.0.0.1:5175">http://127.0.0.1:5175</a></li>
  </ul>
</body>
</html>`
}
