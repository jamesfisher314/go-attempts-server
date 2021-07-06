for source in `ip a | grep inet | grep global | grep -v docker | cut -d'/' -f1 | cut -d't' -f2 | cut -d' ' -f2`; do curl -v --header "X-Forwarded-For: $source" "$source:8080/register?uniquifier=thisisaclearlyinsecurepassword&name=james"; echo $?; echo; echo; done;
for source in `ip a | grep inet | grep global | grep -v docker | cut -d'/' -f1 | cut -d't' -f2 | cut -d' ' -f2`; do curl -v --header "X-Forwarded-For: $source" "$source:8080/register?uniquifier=thisisaninsufficientuniquifier&name=james"; echo $?; echo; echo; done;


