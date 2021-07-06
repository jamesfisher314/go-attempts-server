export addresses=`ip a | grep inet | grep global | grep -v docker | cut -d'/' -f1 | cut -d't' -f2 | cut -d' ' -f2`
echo "Proper registration; expect 204"
for source in $addresses; do curl -s -o /dev/null -w "%{http_code}\n" --header "X-Forwarded-For: $source" "$source:8080/register?uniquifier=thisisaclearlyinsecurepassword&name=james"; echo;  done;
echo "Bad uniqifier; expect 401"
for source in $addresses; do curl -s -o /dev/null -w "%{http_code}\n" --header "X-Forwarded-For: $source" "$source:8080/register?uniquifier=thisisaninsufficientuniquifier&name=james"; echo;  done;
echo "No name; expect 400"
for source in $addresses; do curl -s -o /dev/null -w "%{http_code}\n" --header "X-Forwarded-For: $source" "$source:8080/register?uniquifier=thisisaninsufficientuniquifier"; echo;  done;
echo "No uniquifier; expect 400"
for source in $addresses; do curl -s -o /dev/null -w "%{http_code}\n" --header "X-Forwarded-For: $source" "$source:8080/register?name=james"; echo;  done;
echo "No name or uniquifier; expect 400"
for source in $addresses; do curl -s -o /dev/null -w "%{http_code}\n" --header "X-Forwarded-For: $source" "$source:8080/register"; echo;  done;
echo "Short uniquifier; expect 400"
for source in $addresses; do curl -s -o /dev/null -w "%{http_code}\n" --header "X-Forwarded-For: $source" "$source:8080/register?uniquifier=tooshort&name=james"; echo;  done;
echo "No name, short uniquifier; expect 400"
for source in $addresses; do curl -s -o /dev/null -w "%{http_code}\n" --header "X-Forwarded-For: $source" "$source:8080/register?uniquifier=tooshort"; echo;  done;

echo "Check correct; expect 204"
for source in $addresses; do curl -s -o /dev/null -w "%{http_code}\n" --header "X-Forwarded-For: $source" "$source:8080/check?uniquifier=thisisaclearlyinsecurepassword&name=james"; echo;  done;
echo "Check incorrect; expect 401"
for source in $addresses; do curl -s -o /dev/null -w "%{http_code}\n" --header "X-Forwarded-For: $source" "$source:8080/check?uniquifier=thisisaninsufficientuniquifier&name=james"; echo;  done;
echo "Check no name; expect 401"
for source in $addresses; do curl -s -o /dev/null -w "%{http_code}\n" --header "X-Forwarded-For: $source" "$source:8080/check?uniquifier=thisisaninsufficientuniquifier"; echo;  done;
echo "Check no uniquifier; expect 401"
for source in $addresses; do curl -s -o /dev/null -w "%{http_code}\n" --header "X-Forwarded-For: $source" "$source:8080/check?name=james"; echo;  done;
echo "Check no name or uniquifier; expect 401"
for source in $addresses; do curl -s -o /dev/null -w "%{http_code}\n" --header "X-Forwarded-For: $source" "$source:8080/check"; echo;  done;
echo "Check short uniquifier; expect 401"
for source in $addresses; do curl -s -o /dev/null -w "%{http_code}\n" --header "X-Forwarded-For: $source" "$source:8080/check?uniquifier=tooshort&name=james"; echo;  done;
echo "Check short uniquifier and no name; expect 401"
for source in $addresses; do curl -s -o /dev/null -w "%{http_code}\n" --header "X-Forwarded-For: $source" "$source:8080/check?uniquifier=tooshort"; echo;  done;
