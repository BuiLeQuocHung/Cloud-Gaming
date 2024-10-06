export const stringToUint8Array = (str: string) => {
    const encoder = new TextEncoder();  
    return encoder.encode(str);        
}

export const stringToBase64 = (str: string) => {
    return btoa(str)
}
