let lastMessage = "Aucun message pour le moment";
let lastFile = null;

export const config = {
  api: { bodyParser: false }
};

export default async function handler(req, res) {

  // Route de téléchargement direct
  if (req.method === "GET" && req.url.endsWith("/file")) {
    if (!lastFile) return res.status(404).send("Aucun fichier stocké");

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

  // Route POST
  if (req.method === "POST") {
    try {
      const contentType = req.headers["content-type"] || "";
      if (!contentType.includes("multipart/form-data"))
        return res.status(400).json({ error: "Format non supporté" });

      const boundary = "--" + contentType.split("boundary=")[1];

      // Lire flux brut
      const chunks = [];
      for await (const chunk of req) chunks.push(chunk);
      const buffer = Buffer.concat(chunks);

      // Découper en binaire
      const parts = buffer.split(Buffer.from(boundary));

      for (const part of parts) {
        const headerEnd = part.indexOf("\r\n\r\n");
        if (headerEnd === -1) continue;

        const header = part.slice(0, headerEnd).toString();
        const body = part.slice(headerEnd + 4, part.length - 2); // retirer \r\n

        // Message texte
        if (header.includes('name="message"')) {
          lastMessage = body.toString();
        }

        // Fichier
        if (header.includes("filename=")) {
          const filename = header.match(/filename="(.+?)"/)?.[1] || "fichier.bin";

          lastFile = {
            filename,
            base64: body.toString("base64")
          };
        }
      }

      return res.status(200).json({
        status: "OK",
        receivedMessage: lastMessage,
        receivedFile: lastFile?.filename || null
      });

    } catch (err) {
      return res.status(500).json({ error: "Erreur interne", details: err.message });
    }
  }

  return res.status(405).json({ error: "405" });
};
