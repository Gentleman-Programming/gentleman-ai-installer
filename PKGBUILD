# Maintainer: Gabriel Fagundez <gabrielfagundeznievas@gmail.com>
# Contributor: Gentleman Programming <info@gentlemanprogramming.dev>

pkgname=gentle-ai
pkgver=1.4.1
pkgrel=1
pkgdesc="AI Gentle Stack - Ecosystem configurator for AI coding agents"
arch=('x86_64' 'aarch64')
url="https://github.com/Gentleman-Programming/gentle-ai"
license=('MIT')
provides=("${pkgname}")
conflicts=("${pkgname}")
source_x86_64=("${pkgname}_${pkgver}_linux_amd64.tar.gz::https://github.com/Gentleman-Programming/gentle-ai/releases/download/v${pkgver}/${pkgname}_${pkgver}_linux_amd64.tar.gz")
source_aarch64=("${pkgname}_${pkgver}_linux_arm64.tar.gz::https://github.com/Gentleman-Programming/gentle-ai/releases/download/v${pkgver}/${pkgname}_${pkgver}_linux_arm64.tar.gz")
sha256sums_x86_64=('0703b3fe506d04f46b43d9a0fb44206160bd37794432ba013c5c2cf15f656589')
sha256sums_aarch64=('92896bd7f46b1c34c7b0e3a23dab2e825b8e38a5b4b205006e1bd20a6fb4fa88')

package() {
    # Install binary
    install -Dm755 "${pkgname}" "${pkgdir}/usr/bin/${pkgname}"
    
    # Install man page if exists
    if [ -f "${pkgname}.1" ]; then
        install -Dm644 "${pkgname}.1" "${pkgdir}/usr/share/man/man1/${pkgname}.1"
    fi
}
