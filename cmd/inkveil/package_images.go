package main

import (
 "bytes"
 "fmt"
 "image"
 "image/jpeg"
 "image/png"
 "net/http"
)

// Keep original resolution and transparency. GIF and WebP are left untouched.
func optimizePackageImage(b []byte)([]byte,error){
 kind:=http.DetectContentType(b)
 if kind!="image/jpeg"&&kind!="image/png"{return b,nil}
 cfg,_,err:=image.DecodeConfig(bytes.NewReader(b));if err!=nil{return nil,err}
 if cfg.Width<=0||cfg.Height<=0||int64(cfg.Width)*int64(cfg.Height)>40000000{return nil,fmt.Errorf("image exceeds 40 megapixels")}
 im,_,err:=image.Decode(bytes.NewReader(b));if err!=nil{return nil,err}
 var out bytes.Buffer
 if kind=="image/jpeg"{err=jpeg.Encode(&out,im,&jpeg.Options{Quality:85})}else{enc:=png.Encoder{CompressionLevel:png.BestCompression};err=enc.Encode(&out,im)}
 if err!=nil{return nil,err}
 if out.Len()<len(b){return out.Bytes(),nil};return b,nil
}
