## docker-compose

### override.yml
* Si un fichier docker-compose.override.yml est présent, il est interpreté par défaut sauf si
l'option `-f` est également passé. Si `-f` est passé plusieurs fois, l'ordre compte et le fichier
le plus à droite surcharge ceux à gauche.

### build
* Build tous les services qui ont une section "build".
* Dans la section build `context` indique le path vers un dossier contenant le Dockerfile.
`target` indique quel "stage" du multi-stage build construire.
* Si le service defini aussi une `ìmage` (en plus de `build`), alors le build sera nommé ainsi.

### up
* Build et démarre tous les services. Si le service utilise une `ìmage` qui n'existe pas en local
l'image sera pull sauf si une section `build` est présente, auquelle cas l'image sera build et
sera taggué selon `image`.
* `up` rend les ports réseau indiqués dans le service accessibles depuis le host.

### run
* Adapté pour des commandes adhoc
* Remplace la `command` définie dans le service. L'`entrypoint` est toujours éxécuté mais
la commande passée permet dans ce cas d'ajouter des paramètres supplémentaires.


### depends_on
* Crée une dépendance entre les services, pour influencer l'ordre de démarrage des services
* Ce n'est pas parce que le service est démarré que l'appli (e.g la DB) qui tourne dedans est
up & running. 
* Si on fait `up mysvc` ou `run mysvc` alors tous les services dont dépendent `mysvc` sont aussi démarrés.

### networks
* Par défaut tous les services sont sur le meme réseau, peuvent communiquer entre eux
et sont accessibles avec un hostname égal au nom du service.
* Le reseau est nommé `$(pwd)_default` par défaut.
* https://docs.docker.com/compose/networking/
