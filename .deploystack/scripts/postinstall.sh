terraformDIR="terraform/"
CLIENT=$(terraform -chdir="$terraformDIR" output endpoint)  
CLIENT=${CLIENT/\"/}
CLIENT=${CLIENT/\"/}
echo "Esperando a que el cliente esté activo"

attempt_counter=0
max_attempts=50

until $(curl --output /dev/null --silent --head --fail $CLIENT); do
    if [ ${attempt_counter} -eq ${max_attempts} ];then
    repo=$(git config --get remote.origin.url)
    echo "Intentos máximos alcanzados."
    echo "La solución no se ha instalado correctamente."
    echo 
    echo "Si el problema persiste, envíe una incidencia al repositorio de Github:"
    echo "${repo/.git/}/issues"
    exit 1
    fi

    printf '.'
    attempt_counter=$(($attempt_counter+1))
    sleep 5
done

echo "Éxito, la arquitectura está lista."
echo "Para comprobarlo, visite:"
echo $CLIENT