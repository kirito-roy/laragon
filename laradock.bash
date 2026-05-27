# laradock Bash completion script

_laradock_completion() {
    local cur prev opts
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    opts="list add remove up down restart ssh trust install uninstall"

    case "${prev}" in
        add)
            COMPREPLY=( $(compgen -W "php-82 php-84 php-85" -- ${cur}) )
            return 0
            ;;
        remove)
            local domains=""
            if [ -f "/home/kirito/laragon/sites.txt" ]; then
                domains=$(awk -F'(' '{print $1}' /home/kirito/laragon/sites.txt | sed -e 's|^[^/]*//||' | tr -d ' \r\n\t')
            fi
            COMPREPLY=( $(compgen -W "${domains}" -- ${cur}) )
            return 0
            ;;
        ssh|restart)
            local services="php-82 php-84 php-85 nginx postgres redis rabbitmq node"
            COMPREPLY=( $(compgen -W "${services}" -- ${cur}) )
            return 0
            ;;
    esac

    if [[ ${COMP_CWORD} -eq 1 ]] ; then
        COMPREPLY=( $(compgen -W "${opts}" -- ${cur}) )
        return 0
    fi
}

complete -F _laradock_completion laradock
