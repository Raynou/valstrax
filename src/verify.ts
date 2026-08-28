// Verificación de la firma Ed25519 con la que Discord firma cada petición.
//
// La firma cubre `timestamp + cuerpo crudo`. El cuerpo se lee una sola vez con
// request.text() y se verifica ANTES de parsearlo: parsear y reserializar
// cambia los bytes y rompe la verificación de forma intermitente.

function hexToBytes(hex: string): Uint8Array {
	if (hex.length % 2 !== 0) throw new Error("longitud hex impar");
	const bytes = new Uint8Array(hex.length / 2);
	for (let i = 0; i < bytes.length; i++) {
		const byte = Number.parseInt(hex.slice(i * 2, i * 2 + 2), 16);
		if (Number.isNaN(byte)) throw new Error("carácter hex inválido");
		bytes[i] = byte;
	}
	return bytes;
}

/**
 * Devuelve el cuerpo crudo si la firma es válida, o null si falta alguna
 * cabecera o la firma no corresponde. Nunca lanza.
 */
export async function verifyRequest(
	request: Request,
	publicKey: string,
): Promise<string | null> {
	const signature = request.headers.get("X-Signature-Ed25519");
	const timestamp = request.headers.get("X-Signature-Timestamp");
	if (!signature || !timestamp) return null;

	const body = await request.text();

	try {
		const key = await crypto.subtle.importKey(
			"raw",
			hexToBytes(publicKey),
			{ name: "Ed25519" },
			false,
			["verify"],
		);
		const valid = await crypto.subtle.verify(
			{ name: "Ed25519" },
			key,
			hexToBytes(signature),
			new TextEncoder().encode(timestamp + body),
		);
		return valid ? body : null;
	} catch {
		return null;
	}
}
