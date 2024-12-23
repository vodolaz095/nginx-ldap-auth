curl/private_ok:
	curl --user "abeloy:test321" http://localhost:3000/basic/index.html

curl/private_wrong:
	curl --user "wrong_user:wrong_password" http://localhost:3000/basic/index.html
