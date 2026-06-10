let lastMessage = "Aucun message pour le moment";
let lastFile = null;

export const config = {
  api: { bodyParser: false }
};

export default async function handler(req, res) {
  // Télécharger le fichier
  if (req.method === "GET" && req.url.endsWith("/file")) {
    if (!lastFile) {
      return res.status(404).send("Aucun fichier stocké");
    }

    const buffer = Buffer.from(lastFile.base64, "base64");
    res.setHeader("Content-Type", "application/octet-stream");
    res.setHeader("Content-Disposition", `attachment; filename="${lastFile.filename}"`);
    return res.send(buffer);
  }

  // GET info
  if (req.method === "GET") {
    return res.status(200).json({
      status: "Disponible",
      message: lastMessage,
      file: lastFile ? { filename: lastFile.filename } : null
    });
  }

  // PUT = envoi du fichier brut
  if (req.method === "PUT") {
    try {
      const filename = req.headers["x-filename"] || "fichier.bin";

      const chunks = [];
      for await (const chunk of req) chunks.push(chunk);
      const buffer = Buffer.concat(chunks);

      lastFile = {
        filename,
        base64: buffer.toString("base64")
      };

      return res.status(200).json({
        status: "OK",
        receivedFile: lastFile.filename,
        size: buffer.length
      });
    } catch (e) {
      return res.status(500).json({ error: "Erreur interne", details: e.message });
    }
  }

  return res.status(405).json({ error: "405" });
}
