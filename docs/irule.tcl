when RULE_INIT {
    #set static::sb_debug to 2 if you want to enable logging to troubleshoot this iRule, 1 for informational messages, otherwise set to 0
    set static::sb_debug 2
    if {$static::sb_debug > 1} { log local0. "rule init" }
}

when CLIENTSSL_CLIENTCERT {
    if {$static::sb_debug > 1} {log local0. "In CLIENTSSL_CLIENTCERT"}

    set client_cert [SSL::cert 0]
  
    set serial_id ""
    set spiffe ""
    set log_prefix "[IP::remote_addr]:[TCP::remote_port clientside] [IP::local_addr]:[TCP::local_port clientside]"

    if { [SSL::cert count] > 0 } {
        set spiffe [findstr [X509::extensions [SSL::cert 0]] "Subject Alternative Name" 39 ","]
        if {$static::sb_debug > 1} { log local0. "<$log_prefix>: SAN: $spiffe"}
        set serial_id [X509::serial_number $client_cert]
        if {$static::sb_debug > 1} { log local0. "<$log_prefix>: Serial_ID: $serial_id"}
    }
    if {$static::sb_debug > 1} { log local0.info "here is spiffe: $spiffe" }
       #regexp {^.*\/{[a-zA-Z0-9\-]*}} $spiffe spiffe_result
    set spiffe_result [getfield $spiffe "/" 9]
    log local0. "spiffe_result +++++++++++++ is $spiffe_result"
    set trimspiffe [string trim $spiffe_result]
} 

when CLIENTSSL_HANDSHAKE {
    if { [SSL::extensions exists -type 0] } {
       binary scan [SSL::extensions -type 0] {@9A*} sni_name
       if {$static::sb_debug > 1} { log local0. "sni name: ${sni_name}"}
       regexp {[^.]*} $sni_name sni_result
       log local0. "result is $sni_result"
    }

    # use the ternary operator to return the servername conditionally
    if {$static::sb_debug > 1} { log local0. "sni name: [expr {[info exists sni_name] ? ${sni_name} : {not found} }]"}    
    
    set key [concat $trimspiffe:$sni_result]
    log local0. "here is the key  .... $key"
    log local0.info "target-dg: [class get target-dg]"
    SSL::handshake hold
    if {[class match $key equals "target-dg"] } {
        log local0. "success"
        set gotSNIvalue [class match -value "$key" equals "target-dg"]
        log local0. "value is $gotSNIvalue"
    }
    else {
        log local0. "SNI not in the data group"
        reject
    }
    
    if { $gotSNIvalue eq "allow" } then {
        log local0. "we are good please proceed"
        if {$static::sb_debug > 1} {log local0. "Is the connection authorized: $key"}
        SSL::handshake resume 
    }
    else {
        if {$static::sb_debug > 1} {log local0. "Connection is not authorized: $key"}
        reject
    }
}