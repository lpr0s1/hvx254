let lastMessage = "Aucun message pour le moment";
let lastFile = null;

export const config = {
  api: { bodyParser: false }
};

export default async function handler(req, res) {
  // GET → renvoie le dernier message + fichier
  if (req.method === "GET") {
    return res.status(200).json({
      status: "Disponible",
      message: lastMessage,
      file: lastFile
    });
  }

  // POST → réception message + fichier
  if (req.method === "POST") {
    const contentType = req.headers["content-type"] || "";

    // Lire le flux brut
    const chunks = [];
    for await (const chunk of req) chunks.push(chunk);
    const buffer = Buffer.concat(chunks);

    // Vérification multipart
    if (!contentType.includes("multipart/form-data")) {
      return res.status(400).json({ error: "Format non supporté" });
    }

    // Extraction du boundary
    const boundaryMatch = contentType.match(/boundary=(.+)$/);
    if (!boundaryMatch) {
      return res.status(400).json({ error: "Boundary introuvable" });
    }

    const boundary = "--" + boundaryMatch[1];
    const parts = buffer.toString("binary").split(boundary);

    for (const part of parts) {
      if (!part.includes("Content-Disposition")) continue;

      // Extraction du contenu après les headers
      const [rawHeaders, rawBody] = part.split("\r\n\r\n");
      if (!rawBody) continue;

      const body = rawBody.replace(/\r\n--$/, "");

      // MESSAGE TEXTE
      if (rawHeaders.includes('name="message"')) {
        lastMessage = body.trim();
      }

      // FICHIER
      if (rawHeaders.includes("filename=")) {
        const filename = rawHeaders.match(/filename="(.+?)"/)?.[1];

        if (filename) {
          const binary = Buffer.from(body, "binary");
          lastFile = {
            filename,
            base64: binary.toString("base64")
          };
        }
      }
    }

    return res.status(200).json({
      status: "OK",
      receivedMessage: lastMessage,
      receivedFile: lastFile ? lastFile.filename : null
    });
  }

  // Méthode non autorisée
  return res.status(405).json({ error: "405" });
}
