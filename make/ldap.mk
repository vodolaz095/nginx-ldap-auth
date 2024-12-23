ldap/edit:
	ldapvi --discover --host="127.0.0.1:1389" --user="cn=admin,dc=vodolaz095,dc=ru" --password="someRandomPasswordToMakeHackersSad22223338888"

ldap/upload:
	ldapvi --in --host="127.0.0.1:1389" --user="cn=admin,dc=vodolaz095,dc=ru" --password="someRandomPasswordToMakeHackersSad22223338888" contrib/seed.ldif
