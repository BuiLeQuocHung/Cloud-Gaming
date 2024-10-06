export const uint8ArrayToString = (str) => { 
    const decoder = new TextDecoder('utf-8');
    return decoder.decode(str);
}

export const base64ToString = (str) => {
    return atob(str)
}