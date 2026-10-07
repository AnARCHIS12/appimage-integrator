# Aurémi — AppImage Integrator

**Aurémi** est un petit gestionnaire qui transforme un double-clic sur un fichier AppImage en installation propre. Le gestionnaire reste invisible au quotidien ; seule l’installation initiale utilise une petite boîte de dialogue.

![Logo Aurémi](assets/auremi-logo.png)

Il fonctionne avec les bureaux qui respectent les standards Linux/Freedesktop : KDE Plasma, GNOME, Cinnamon, XFCE, MATE, LXQt et la plupart des autres environnements.

## Ce qu’il fait

Lorsque vous double-cliquez sur une AppImage, AppImage Integrator :

1. vérifie qu’il s’agit réellement d’une AppImage ;
2. lit son nom et son icône sans lancer l’application ;
3. en conserve une copie exécutable dans votre dossier personnel ;
4. crée une entrée dans le menu des applications ;
5. crée un raccourci sur le Bureau ;
6. affiche une notification lorsque l’installation est terminée.

Le fichier téléchargé d’origine n’est jamais supprimé. L’AppImage n’est jamais lancée pendant son installation.

## Installation par double-clic — la méthode la plus simple

1. Téléchargez l’archive adaptée à votre PC : `Auremi-Installer-linux-amd64.tar.gz` pour la majorité des ordinateurs, ou `Auremi-Installer-linux-arm64.tar.gz` pour une machine ARM.
2. Extrayez l’archive avec votre gestionnaire de fichiers.
3. Double-cliquez sur `Auremi-Installer-linux-amd64` ou `Auremi-Installer-linux-arm64`.
4. Cliquez sur **Installer**.

L’installateur affiche une boîte de dialogue native avec KDialog sur KDE, Zenity sur GNOME et les bureaux compatibles, ou une notification en dernier recours. Il installe uniquement dans votre compte et ne demande pas de mot de passe administrateur.

## Installation en ligne de commande

### Pour votre compte uniquement — recommandé

Cette méthode ne demande pas `sudo` :

```bash
chmod +x appimage-integrator-linux-amd64
./appimage-integrator-linux-amd64 setup --user
```

Le gestionnaire est copié dans `~/.local/bin/` et enregistré uniquement pour votre compte.

### Pour tous les utilisateurs du PC

Cette commande modifie `/usr/local/bin`, `/usr/share` et `/etc/xdg`. Elle demande donc les droits administrateur :

```bash
chmod +x appimage-integrator-linux-amd64
sudo ./appimage-integrator-linux-amd64 setup --system
```

Après cette installation unique, chaque utilisateur peut intégrer ses AppImage sans `sudo`. Les applications restent séparées dans le dossier personnel de chaque utilisateur.

## Utilisation

Après l’installation du gestionnaire, double-cliquez simplement sur un fichier portant l’extension `.AppImage`.

L’application apparaîtra ensuite :

- dans le menu des applications ;
- sur le Bureau ;
- dans `~/.local/share/appimage-integrator/apps/`.

Le premier clic installe l’AppImage. Pour la lancer, utilisez ensuite son entrée dans le menu ou son raccourci sur le Bureau.

### Installation manuelle

```bash
appimage-integrator integrate MonApplication.AppImage
```

Sans raccourci sur le Bureau :

```bash
appimage-integrator integrate --no-desktop MonApplication.AppImage
```

### Voir les applications gérées

```bash
appimage-integrator list
```

### Désinstaller une application

Repérez d’abord son identifiant avec `list`, puis :

```bash
appimage-integrator remove appimage-mon-application
```

Cette commande retire uniquement la copie gérée, l’icône et les lanceurs. Le fichier AppImage téléchargé à l’origine reste intact.

## Compatibilité Linux

Les binaires fournis sont autonomes et ne dépendent ni de GTK, ni de Qt, ni d’une distribution particulière. Deux versions sont produites : `linux-amd64` pour les PC Intel/AMD 64 bits et `linux-arm64` pour les machines ARM 64 bits. Comme ils sont compilés statiquement, ils fonctionnent aussi bien sur les distributions utilisant glibc que sur celles utilisant musl.

Le fichier `dist/SHA256SUMS` permet de vérifier que les binaires n’ont pas été modifiés :

```bash
cd dist
sha256sum -c SHA256SUMS
```

La commande facultative `unsquashfs`, fournie par le paquet `squashfs-tools`, permet d’extraire le vrai nom et l’icône embarquée. Si elle n’est pas disponible, l’installation fonctionne quand même avec le nom du fichier et une icône générique.

Installation facultative de `squashfs-tools` :

```bash
# Debian, Ubuntu, Linux Mint
sudo apt install squashfs-tools

# Fedora
sudo dnf install squashfs-tools

# Arch, Manjaro
sudo pacman -S squashfs-tools

# openSUSE
sudo zypper install squashfs

# Alpine
sudo apk add squashfs-tools
```

Les utilitaires `update-mime-database`, `update-desktop-database`, `gtk-update-icon-cache`, `kbuildsycoca6` et `notify-send` sont utilisés lorsqu’ils sont présents, mais aucun n’est obligatoire.

## Sécurité

AppImage Integrator réduit les risques pendant l’installation :

- il refuse les liens symboliques et les fichiers qui ne sont pas des AppImage ;
- il ne lance jamais l’AppImage pendant l’analyse ou l’installation ;
- il utilise `unsquashfs` pour lire les métadonnées lorsque cet outil est disponible ;
- il écrit les fichiers de manière atomique ;
- il ne demande jamais `sudo` pour intégrer une application dans un compte utilisateur ;
- il ne supprime jamais le téléchargement d’origine.

Cela ne prouve toutefois pas qu’une AppImage est digne de confiance. Téléchargez toujours vos applications depuis leur site ou dépôt officiel et vérifiez les sommes SHA-256 publiées lorsque celles-ci existent.

## Emplacements utilisés

Installation utilisateur du gestionnaire :

- binaire : `~/.local/bin/appimage-integrator`
- gestionnaire de fichiers : `~/.local/share/applications/appimage-integrator-handler.desktop`
- définition MIME : `~/.local/share/mime/packages/appimage-integrator.xml`

Applications intégrées :

- AppImage et métadonnées : `~/.local/share/appimage-integrator/apps/`
- lanceurs : `~/.local/share/applications/`
- icônes : `~/.local/share/icons/hicolor/`
- raccourcis : dossier Bureau défini par `XDG_DESKTOP_DIR`

Les variables standard `XDG_DATA_HOME` et `XDG_CONFIG_HOME` sont respectées.

## Compiler depuis les sources

Go 1.23 ou plus récent est nécessaire :

```bash
make test
make
```

Les deux binaires statiques sont créés dans `dist/`.

Pour une autre architecture :

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags="-s -w" -o dist/appimage-integrator-linux-arm64 .
```

## Limites connues

- Les AppImage de type 1 et 2 sont reconnues, mais l’extraction automatique de l’icône dépend de `unsquashfs` et d’un contenu SquashFS lisible.
- Certains gestionnaires de fichiers demandent une première fois avec quelle application ouvrir une AppImage. Choisissez **Installer une AppImage** et cochez l’option permettant de mémoriser ce choix.
- Une installation globale rend le gestionnaire disponible à tous, mais chaque AppImage est volontairement installée dans le compte de l’utilisateur qui clique dessus. Cela évite de demander le mot de passe administrateur à chaque application.

## Licence

MIT — voir `LICENSE`.
