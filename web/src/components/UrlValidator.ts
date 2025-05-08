import validator from "validator";

export function isValidUrl(url: string): boolean {
    var options = { 
        protocols: ['http','https'], 
        require_tld: true, 
        require_protocol: true, 
        require_host: true, 
        require_port: false, 
        require_valid_protocol: true, 
        allow_underscores: false, 
        host_whitelist: undefined, 
        host_blacklist: undefined, 
        allow_trailing_dot: false, 
        allow_protocol_relative_urls: false, 
        allow_fragments: true, 
        allow_query_components: true, 
        disallow_auth: false, 
        validate_length: true }

    return validator.isURL(url, options);
}