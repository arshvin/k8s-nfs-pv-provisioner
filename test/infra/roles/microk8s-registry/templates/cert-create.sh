#/bin/bash
openssl req -new -config {{ role_tmp_directory }}/openssl.conf -x509 -days 365 -extensions x509 > {{ proxy_cert_file }}