{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  buildInputs = with pkgs; [
    chafa
    # Go compiler and tools
    go
    gopls
    gotools
    go-tools
    
    # Tesseract OCR and dependencies
    tesseract
    # tesseract-lang # Language data for Tesseract
    leptonica      # Image processing library used by Tesseract
    
    # Common build tools
    gcc
    pkg-config
    
    # For image processing
    libpng
    libjpeg
    libtiff
  ];

  shellHook = ''
    echo "Nix shell for grepShot - Go application with Tesseract OCR"
    echo "Tesseract OCR version: $(${pkgs.tesseract}/bin/tesseract --version | head -n 1)"
    echo "Go version: $(${pkgs.go}/bin/go version)"
    
    # Set PKG_CONFIG_PATH to find Tesseract and Leptonica
    export PKG_CONFIG_PATH="${pkgs.tesseract}/lib/pkgconfig:${pkgs.leptonica}/lib/pkgconfig:$PKG_CONFIG_PATH"
    
    # Set CGO_CFLAGS and CGO_LDFLAGS for gosseract
    export CGO_CFLAGS="-I${pkgs.tesseract}/include/tesseract -I${pkgs.leptonica}/include"
    export CGO_LDFLAGS="-L${pkgs.tesseract}/lib -L${pkgs.leptonica}/lib -ltesseract -lleptonica"
    
  '';
}
