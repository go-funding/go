with
  domains_to_insert as (
    select
      substr(host, 5) as host
    from
      domains
    where
      host like 'www.%'
      and substr(host, 5) not in (select host from domains)
  )
insert into domains (host) select host from domains_to_insert;
