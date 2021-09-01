package server

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type gitinfo struct {
	T string `json:"tag_name"`
	N string `json:"name"`
	U string `json:"tarball_url"`
	A []struct {
		N string `json:"name"`
		U string `json:"browser_download_url"`
	} `json:"assets"`
}

const gitFS, gitPlugs, tmpPlugins, clearINI = "damiva/ForkServer", "damiva/ForkServerPlugs", "plugins.tar.gz", "clear.ini"

var instaNew bool

func (i *gitinfo) getAss() string {
	for _, a := range i.A {
		if strings.Contains(a.N, runtime.GOOS+"-"+runtime.GOARCH) {
			return a.U
		}
	}
	return ""
}
func gitInfo(s string) (i gitinfo, e error) {
	var r *http.Response
	if r, e = httpGet("https://api.github.com/repos/"+s+"/releases/latest", http.Header{"Accespt": {"application/vnd.github.v3+json"}}); e == nil {
		defer r.Body.Close()
		defer ioutil.ReadAll(r.Body)
		if e = json.NewDecoder(r.Body).Decode(&i); e != nil {
			e = errors.New("Parsing GitHub " + s + " error: " + e.Error())
		}
	}
	return
}
func download(src, dst string) (e error) {
	var f *os.File
	if f, e = os.Create(dst); e == nil {
		defer f.Close()
		var r *http.Response
		if r, e = httpGet(src, nil); e == nil {
			defer r.Body.Close()
			_, e = io.Copy(f, r.Body)
		}
	}
	return
}
func extractPlugins(n string) (e error) {
	if f, e := os.Open(n); e == nil {
		defer f.Close()
		if z, e := gzip.NewReader(f); e == nil {
			defer z.Close()
			var h *tar.Header
			t := tar.NewReader(z)
			for h, e = t.Next(); e == nil; h, e = t.Next() {
				if p := strings.SplitN(h.Name, "/", 2); h.Typeflag == tar.TypeReg && len(p) == 2 && p[1] == clearINI {
					var b []byte
					if i, e := t.Read(b); e != nil {
						Error("Reading", tmpPlugins, ", file", h.Name, "error:", e)
					} else if i > 0 {
						for _, n := range bytes.Split(b, []byte{'\n'}) {
							fn := strings.TrimSpace(string(n))
							if i, e := os.Stat(fn); e != nil {
								Info("Clearing plugins error:", e)
							} else if i.IsDir() {
								if e = os.RemoveAll(fn); e != nil {
									Info("Clearing plugins error:", e)
								}
							} else if e = os.Remove(fn); e != nil {
								Info("Clearing plugins error:", e)
							}
						}
					}
					e = nil
					break
				}
			}
			t = tar.NewReader(z)
			for h, e = t.Next(); e == nil; h, e = t.Next() {
				if p := strings.SplitN(h.Name, "/", 2); len(p) == 2 {
					switch h.Typeflag {
					case tar.TypeDir:
						e = os.MkdirAll(filepath.Join(pthPlugs, p[1]), 0777)
					case tar.TypeReg:
						if p[1] != clearINI {
							if i, e := os.Create(p[1]); e == nil {
								_, e = io.Copy(i, t)
								i.Close()
							}
						}
					}
				}
			}
			if e == io.EOF {
				e = nil
			}
		}
	}
	return
}
func updateFS(justCheck bool) (nv bool) {
	var (
		i gitinfo
		e error
		t = "Check update of " + Name + ":"
	)
	if i, e = gitInfo(gitFS); e != nil {
		Error(t, e)
	} else if nv = i.T != "" && len(i.A) > 0 && i.T != Vers; !nv {
		Info(t, "there is no update.")
	} else if i.U = i.getAss(); i.U == "" {
		Error(t, "there is no distrib for", runtime.GOOS, "/", runtime.GOARCH, "!")
		nv = false
	} else if justCheck {
		return
	} else if e = download(i.U, exec+".new"); e == nil {
		mutex.Lock()
		if e = os.Chmod(exec+".new", 0777); e == nil {
			if e = os.Rename(exec, exec+".old"); e == nil {
				if e = os.Rename(exec+".new", exec); e == nil {
					go Restart()
					Warning("Restarting", Name, "...")
				} else {
					os.Rename(exec+".old", exec)
				}
			}
		}
		mutex.Unlock()
	}
	if e != nil {
		Error(e)
	}
	return
}
func updatePS(justCheck bool) (nv bool) {
	t, rm := "Check update of plugins:", false
	if i, e := gitInfo(gitPlugs); e != nil {
		Error(t, e)
	} else if nv = i.U != "" && i.T != "" && i.T != sets.PlugsTag; !nv {
		Info(t, "there is no update.")
	} else if nm := strings.Split(i.N, " "); strings.TrimSuffix(nm[len(nm)-1], "+") > Vers {
		Info(t, "there is no update for your ForkServer version.")
	} else if justCheck {
		return
	} else if e = download(i.U, tmpPlugins); e != nil {
		Error(e)
	} else if e = extractPlugins(tmpPlugins); e != nil {
		Error(e)
		rm = true
	} else {
		sets.PlugsTag, rm = i.T, true
		if e = sets.put(); e != nil {
			Error(e)
		}
		Warning("Plugins has been updated to v.", sets.PlugsTag)
	}
	if rm {
		os.Remove(tmpPlugins)
	}
	return
}
func checkUpdate(nv bool, auto time.Duration) {
	if auto > 0 {
		if !nv {
			updateFS(false)
		}
		updatePS(false)
		t := time.Tick(auto)
		go func() {
			for range t {
				instaNew = updateFS(true) || updatePS(true)
			}
		}()
	}
}
func update(w http.ResponseWriter, r *http.Request) {
	t := [3]string{"Updated!", " Restarting...", "There are no updates!"}
	if r.FormValue("w_lang") == "ru" {
		t = [3]string{"Обновлено!", " Перезапускатеся...", "Нет обновлений!"}
	}
	instaNew = false
	if updateFS(false) {
		w.Write([]byte(`{"notify":"` + t[0] + t[1] + `","cmd":"reload(2);`))
	} else if updatePS(false) {
		w.Write([]byte(`{"notify":"` + t[0] + `","cmd":"reload(0.001);`))
	} else {
		w.Write([]byte(`{"notify":"` + t[2] + `","cmd":"stop();`))
	}
}
func httpGet(u string, h http.Header) (r *http.Response, e error) {
	var q *http.Request
	if q, e = http.NewRequest("GET", u, nil); e == nil {
		for k, vs := range h {
			for _, v := range vs {
				q.Header.Add(k, v)
			}
		}
		c := httpClient()
		if r, e = c.Do(q); e == nil {
			if r.StatusCode != 200 {
				ioutil.ReadAll(r.Body)
				r.Body.Close()
				e = errors.New(u + " returned: " + r.Status)
			}
		}
	}
	return
}
