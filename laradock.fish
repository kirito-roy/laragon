# laradock Fish completion script

# Disable file completion by default
complete -c laradock -f

# Main subcommands
complete -c laradock -n "not __fish_seen_subcommand_from list add remove up down restart ssh trust install uninstall" -a list -d "List active projects and their PHP versions"
complete -c laradock -n "not __fish_seen_subcommand_from list add remove up down restart ssh trust install uninstall" -a add -d "Register a new local project site"
complete -c laradock -n "not __fish_seen_subcommand_from list add remove up down restart ssh trust install uninstall" -a remove -d "Deregister a project site"
complete -c laradock -n "not __fish_seen_subcommand_from list add remove up down restart ssh trust install uninstall" -a up -d "Start all LaraDock containers in background"
complete -c laradock -n "not __fish_seen_subcommand_from list add remove up down restart ssh trust install uninstall" -a down -d "Stop and remove all containers"
complete -c laradock -n "not __fish_seen_subcommand_from list add remove up down restart ssh trust install uninstall" -a restart -d "Restart all or a specific service"
complete -c laradock -n "not __fish_seen_subcommand_from list add remove up down restart ssh trust install uninstall" -a ssh -d "Shell instantly into a running container"
complete -c laradock -n "not __fish_seen_subcommand_from list add remove up down restart ssh trust install uninstall" -a trust -d "Trust the local Certificate Authority on the host"
complete -c laradock -n "not __fish_seen_subcommand_from list add remove up down restart ssh trust install uninstall" -a install -d "Install global CLI, man pages, and completions"
complete -c laradock -n "not __fish_seen_subcommand_from list add remove up down restart ssh trust install uninstall" -a uninstall -d "Wipe LaraDock global CLI and configurations"

# Subcommand completions
# add php version suggestion
complete -c laradock -n "__fish_seen_subcommand_from add" -a "php-82 php-84 php-85" -d "PHP Container Version"

# ssh / restart service suggestion
complete -c laradock -n "__fish_seen_subcommand_from ssh restart" -a "php-82 php-84 php-85 nginx postgres redis rabbitmq node" -d "LaraDock Service"

# remove registered domains suggestion
complete -c laradock -n "__fish_seen_subcommand_from remove" -a "(awk -F'(' '{print \$1}' /home/kirito/laragon/sites.txt | sed -e 's|^[^/]*//||' | tr -d ' \r\n\t')" -d "Registered Domain"
