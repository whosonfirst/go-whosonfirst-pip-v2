package test

import {
    "github.com/mdwhatcott/gounit"
    "strings"
    "testing"
    "wof-pip-server"
}

func Test(t *testing.T) {
    var ps *wof-pip-server

    f := NewFixture("PIP Server", t)
    defer f.Run()

    f.Setup(func() {
        ps = wof-pip-server{}
        # init ps
    })
    
    f.Test("extras=name:", func() {
        result := ps.GetByLatLon(36.1, 140.08, "extras=name:")
        
        f.So("result will have places [Japan, Ibaraki, Tsukuba] ", len(result["places"]), ShouldEqual, 3)
        for _, place := range result["places"] {
            exist := false
            for k, v := range place {
                if strings.HasPrefix(k, "name:") {
                   exist = true
                }

            }
            f.So("Each place will have name:", exist, ShouldBeTrue, ok)
       }
    })

    f.Test("extras=name::jpn_x_preferred", func() {
        result := ps.GetByLatLon(36.1, 140.08, "extras=name:")
        
        f.So("result will have places [Japan, Ibaraki, Tsukuba] ", len(result["places"]), ShouldEqual, 3)
        for _, place := range result["places"] {
            exist := false
            for k, v := range place {
                if strings.HasPrefix(k, "name:jpn_x_preferred") {
                   exist = true
                }

            }
            f.So("Each place will have name::jpn_x_preferred", exist, ShouldBeTrue, ok)
       }
    })
}
