let lastMessage = "Aucun message pour le moment";
let lastFile = null;

export const config = {
  api: { bodyParser: false }
};

export default async function handler(req, res) {
  // Route spéciale pour télécharger le fichier directement
  if (req.method === "GET" && req.url.endsWith("/file")) {
    if (!lastFile) {
      return res.status(404).send("Aucun fichier stocké");
    }

    const buffer = Buffer.from(lastFile.base64, "base64");

    res.setHeader("Content-Type", "application/octet-stream");
    res.setHeader("Content-Disposition", `attachment; filename="${lastFile.filename}"`);
    return res.send(buffer);
  }

  // Route GET normale
  if (req.method === "GET") {
    return res.status(200).json({
      status: "Disponible",
      message: lastMessage,
      file: lastFile
    });
  }

  // Route POST pour recevoir message + fichier
  if (req.method === "POST") {
    try {
      const contentType = req.headers["content-type"] || "";
      if (!contentType.includes("multipart/form-data")) {
        return res.status(400).json({ error: "Format non supporté" });
      }

      // Lire flux brut
      const chunks = [];
      for await (const chunk of req) chunks.push(chunk);
      const buffer = Buffer.concat(chunks);

      // Boundary
      const boundary = contentType.split("boundary=")[1];
      const delimiter = "--" + boundary;
      const parts = buffer.toString("latin1").split(delimiter);

      for (const part of parts) {
        if (!part.includes("Content-Disposition")) continue;

        const [rawHeaders, rawBody] = part.split("\r\n\r\n");
        if (!rawBody) continue;

        const body = rawBody.replace(/\r\n--$/, "");

        // Message texte
        if (rawHeaders.includes('name="message"')) {
          lastMessage = body;
        }

        // Fichier
        if (rawHeaders.includes("filename=")) {
          const filenameMatch = rawHeaders.match(/filename="(.+?)"/);
          const filename = filenameMatch ? filenameMatch[1] : "fichier.bin";

          const binaryBuffer = Buffer.from(body, "latin1");

          lastFile = {
            filename,
            base64: binaryBuffer.toString("base64")
          };
        }
      }

      return res.status(200).json({
        status: "OK",
        receivedMessage: lastMessage,
        receivedFile: lastFile ? lastFile.filename : null
      });

    } catch (err) {
      return res.status(500).json({
        error: "Erreur interne",
        details: err.message
      });
    }
  }

  return res.status(405).json({ error: "405" });
}
