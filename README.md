# IrisPropera

IrisPropera est la version go du back-end de Propera. Le premier back-end a été écrit en PHP avec le framework Laravel 5.5.

Afin d'accélérer les traitements en particulier sur les imports de données issues d'IRIS, le back-end a été transcrit en Go. Cela permet également de régler les problèmes de tests unitaires qu'il n'était pas possible de faire compte tenu de la manière dont les token ont été implémentés dans la première version du back-end.

IrisPropera utilise le "framework" iris sous go.

Le système de configuration est spécifique au projet et utilise le package yaml-v2.


## Structure du projet

Le projet s'inspire d'une structure MVC :

* `main.go` fichier principal lançant le serveur
* `config.yml` fichier de configuration de la base de données et des tests unitaires
* `actions/` package contenant l'ensemble des handlers et des fichiers de test correspondants ainsi que le fichier de routing
* `models/`modèles/tables de la base de données contenant les requêtes en PostgreSQL permettant de fournir les résultats aux actions
* `config/` configuration d'IrisPropera et de lancement de la base de données

## Organisation des tests

Les tests respectent globalement la philosophie générale de Go consistant à tester unitairement chaque fichier de chaque package grâce à un fichier test situé dans le même répertoire.

Compte tenu de son rôle particulier et de la difficulté de faire des tests unitaires, le fichier de lancement `main.go` n'est associé à aucun fichier de test.

### Utilisation d'une base de données de tests

Les tests des handlers ne sont pas réalisés avec des mockers de la base de données pour s'assurer que les requêtes SQL des modèles fonctionnent également correctement.

Pour s'assurer que les tests ne seront pas perturbés par la base de données, une base de données de test propera3_test, sauvegardée par ailleurs, est restaurée à chaque lancement de test. De même, compte tenu de la protection des entrées de l'API par des middlewares, des connexions doivent être réalisées préalablement à tout test.

Une copie de la base de test a été effectuée. La séquence d'utilitaires PostgreSQL utilisée est la suivante :
```
pg_dump -Fc -w -U postgres -f [db_dump] -d [base production]
create_db -O postgres [db_dump]
pg_restore -cOU postgres -d [base test] [db_dump]
```

Le fichier de dump est stocké localement mais non inclus dans le git repository. Son emplacement est stocké dans une variable système ainsi que le mot de passe d'accès à la base de données.

La fonction `TestCommons` du fichier `commons_test.go` implémente donc la récupération de la configuration en particulier pour la localisation du dump de la base et du nom de la base de test. Elle lance la troisième commande `pg_restore` et ignore les erreurs non `FATAL` qui peuvent être liées au fait que les tests ont altéré la structure de la base de test, par exemple en créant des tables provisoires pour les imports en batch.

Tous les tests des handlers doivent donc appeler cette fonction avant de lancer leurs propres tests.

### Structure des tests des handlers

Les tests des handlers ne contiennent qu'une seule fonction de test principale qui appelle préalablement TestCommons pour s'assurer de l'initialisation de la base de données et de la disponibilité d'utilisateurs connectés et pour lancer des sous-tests pour les différentes fonctions.

Les sous-fonctions correspondent à une action et donc à un point d'entrée de l'API.

Ces sous-tests ne sont pas directement accessibles et n'ont pas le format reconnu par Go pour réaliser des fonctions de test, le test de chaque fonction nécessitant préalablement une initialisation de la base de données et de la connexion des utilisateurs.

Pratiquement tous les tests ont la même forme et respectent la philosophie générale de Go à savoir un tableau de cas de tests pour chaque fonction et une vérification du retour de la requête. Les assertions sont faites sous une forme basique mentionnant toutefois systématiquement la référence du cas de tests pour un débogage plus rapide.

Les requêtes et le décodage utilisent le système de test du framework IRIS. Cependant, les données sont réinterprétées en Go classique pour faire les assertions et pour afficher les erreurs.

## Différences par rapport à la version PHP du back-end

Quelques routes ont été modifiées par rapport à la première version du back end et nécessiteront une correction dans le front end. Elles sont documentées par des commentaires dans le fichier `actions\routes.go`.

Afin de réduire le temps nécessaire pour l'affichage des pages qui est surtout lié à la latence du réseau, des requêtes sont groupées afin qu'une page du front end ne fasse qu'une requête GET pour l'ensemble de son contenu à chaque fois que cela est possible.
