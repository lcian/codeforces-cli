package client

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"syscall"

	"github.com/fatih/color"
	"github.com/lcian/codeforces-cli/cookiejar"
	"github.com/lcian/codeforces-cli/util"
	"golang.org/x/crypto/ssh/terminal"
)

// genFtaa generate a random one
func genFtaa() string {
	return util.RandString(18)
}

// genBfaa generate a bfaa
func genBfaa() string {
	return "f1b3f18c715565b589b7823cda7448ce"
}

// ErrorNotLogged not logged in
var ErrorNotLogged = "Not logged in"

// findHandle if logged return (handle, nil), else return ("", ErrorNotLogged)
func findHandle(body []byte) (string, error) {
	reg := regexp.MustCompile(`handle = "([\s\S]+?)"`)
	tmp := reg.FindSubmatch(body)
	if len(tmp) < 2 {
		return "", errors.New(ErrorNotLogged)
	}
	return string(tmp[1]), nil
}

func findCsrf(body []byte) (string, error) {
	reg := regexp.MustCompile(`csrf='(.+?)'`)
	tmp := reg.FindSubmatch(body)
	if len(tmp) < 2 {
		return "", errors.New("Cannot find csrf")
	}
	return string(tmp[1]), nil
}

// Prompt the user for the cf_clearance Cloudflare cookie and store it in the cookie jar
func (c *Client) GetCloudflareCookieFromUser() (err error) {
	color.Cyan("Cloudflare cookie required")
	color.Cyan("Inspect the cookies in your browser (e.g. in Chrome, on the Codeforces home page: Right click > Inspect > Application > Cookies > https://codeforces.com) and copy the required cookie")
	fmt.Printf("cf_clearance: ")
	cf_clearance := util.ScanlineTrim()
	cf_clearance_cookie := http.Cookie{
		Name:   "cf_clearance",
		Value:  cf_clearance,
		Path:   "/",
		Domain: "codeforces.com",
	}
	url, _ := url.Parse("codeforces.com")
	c.client.Jar.SetCookies(url, append(c.client.Jar.Cookies(url), &cf_clearance_cookie))
	return nil
}

// Login codeforces with handler and password
func (c *Client) Login() (err error) {
	color.Cyan("Login %v...\n", c.HandleOrEmail)

	password, err := c.DecryptPassword()
	if err != nil {
		return
	}

	jar, _ := cookiejar.New(nil)
	c.client.Jar = jar
	body, err := util.GetBody(c.client, c.host+"/enter")
	if err != nil {
		return
	}

	csrf, err := findCsrf(body)
	if err != nil {
		color.Red(err.Error())
		/* One of the failure scenarios for findCsrf is when our request for "/enter" is redirected to a Cloudflare challenge page
		 * Let's get the cf_clearance cookie that Cloudflare sets when a challenge is passed and retry
		 */
		err = c.GetCloudflareCookieFromUser()
		if err != nil {
			color.Red(err.Error())
			return
		}
		body, err = util.GetBody(c.client, c.host+"/enter")
		if err != nil {
			return
		}
		csrf, err = findCsrf(body)
		if err != nil {
			return
		}
	}

	ftaa := genFtaa()
	bfaa := genBfaa()

	body, err = util.PostBody(c.client, c.host+"/enter", url.Values{
		"csrf_token":    {csrf},
		"action":        {"enter"},
		"ftaa":          {ftaa},
		"bfaa":          {bfaa},
		"handleOrEmail": {c.HandleOrEmail},
		"password":      {password},
		"_tta":          {"176"},
		"remember":      {"on"},
	})
	if err != nil {
		return
	}

	handle, err := findHandle(body)
	if err != nil {
		return
	}

	c.Ftaa = ftaa
	c.Bfaa = bfaa
	c.Handle = handle
	c.Jar = jar
	color.Green("Succeed!!")
	color.Green("Welcome %v~", handle)
	return c.save()
}

func createHash(key string) []byte {
	hasher := md5.New()
	hasher.Write([]byte(key))
	return hasher.Sum(nil)
}

func encrypt(handle, password string) (ret string, err error) {
	block, err := aes.NewCipher(createHash("glhf" + handle + "233"))
	if err != nil {
		return
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return
	}
	text := gcm.Seal(nonce, nonce, []byte(password), nil)
	ret = hex.EncodeToString(text)
	return
}

func decrypt(handle, password string) (ret string, err error) {
	data, err := hex.DecodeString(password)
	if err != nil {
		err = errors.New("Cannot decode the password")
		return
	}
	block, err := aes.NewCipher(createHash("glhf" + handle + "233"))
	if err != nil {
		return
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return
	}
	nonceSize := gcm.NonceSize()
	nonce, text := data[:nonceSize], data[nonceSize:]
	plain, err := gcm.Open(nil, nonce, text, nil)
	if err != nil {
		return
	}
	ret = string(plain)
	return
}

// DecryptPassword get real password
func (c *Client) DecryptPassword() (string, error) {
	if len(c.Password) == 0 || len(c.HandleOrEmail) == 0 {
		return "", errors.New("You have to configure your handle and password by `cf config`")
	}
	return decrypt(c.HandleOrEmail, c.Password)
}

// ConfigLogin configure handle and password
func (c *Client) ConfigLogin() (err error) {
	if c.Handle != "" {
		color.Green("Current user: %v", c.Handle)
	}
	color.Cyan("Configure handle/email and password")
	color.Cyan("Note: The password is invisible, just type it correctly.")

	fmt.Printf("handle/email: ")
	handleOrEmail := util.ScanlineTrim()

	password := ""
	if terminal.IsTerminal(int(syscall.Stdin)) {
		fmt.Printf("password: ")
		bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			fmt.Println()
			if err.Error() == "EOF" {
				fmt.Println("Interrupted.")
				return nil
			}
			return err
		}
		password = string(bytePassword)
		fmt.Println()
	} else {
		color.Red("Your terminal does not support the hidden password.")
		fmt.Printf("password: ")
		password = util.Scanline()
	}

	c.HandleOrEmail = handleOrEmail
	c.Password, err = encrypt(handleOrEmail, password)
	if err != nil {
		return
	}
	return c.Login()
}
